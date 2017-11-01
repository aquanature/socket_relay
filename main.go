package main

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"errors"
	"github.com/aquanature/socket_relay/config"
	"github.com/aquanature/socket_relay/sbrp"
	"strings"
)

var cfg config.Cfg

type RelayConn struct {
	conn     *net.Conn // Reference to the connection used by the host
	port     int       // Port that clients connect to for reaching this host
	name     string    // Human readable name that is configurable
	useSbrp  bool      // Tells the relay whether to use the "Stephane Bedard Relay Protocol" or SBRP with this host
	id       int
	idParent int
	op       sbrp.Op
}

type DataPack struct {
	data     []byte
	length   int
	idSource int
}

// Main manages:
// - The global channels
// - The master list of currently managed hosts
// - Manage any exit channels and signals
// - Manage the graceful exiting of all goroutines, especially the closing of all OS/System connections.
func main() {
	mainLoop()
}

func mainLoop() {
	fmt.Println("Starting Relay")

	cfg.Init()

	// Map used for managing the various hosts connected to the relay
	relayConns := make(map[int]RelayConn)

	opChan := make(chan sbrp.Op)
	hostListChan := make(chan RelayConn)

	l, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.HostPort))
	if err != nil {
		fmt.Errorf("cannot listen on port " + strconv.Itoa(cfg.HostPort) + ": likely cause port in use")
		return
	} else {
		// waitForHostConnections() is launched in it's own goroutine so as to not block the channel message management
		go waitForHostConnections(l, opChan, hostListChan)
	}

	defer exit(l, relayConns)

	// A loop that blocks until it received a message on an assigned channel and reacts to it.
	connId := 0
	op := sbrp.Loop
	for op == sbrp.Loop {
		select {
		case rc := <-hostListChan:
			switch rc.op {
			case sbrp.Add:
				connId++
				rc.id = connId
				relayConns[connId] = rc
			case sbrp.Remove:
				delete(relayConns, rc.id)
			}
		case op = <-opChan:
			switch op {
			case sbrp.Quit:
				fmt.Println("Exit requested with code " + op.ToStr())
				break
			default:
				// If it's an unrecognized operation, continue the update loop
				op = sbrp.Loop
			}
		}
	}
	fmt.Println("Exit with code " + op.ToStr())
}

// Final deferred cleanup that verifies that all connections are closed before Go destroys the thread
// Important since the OS often will not release a port until it has a system timeout or it is explicitly closed.
func exit(l net.Listener, relayConnList map[int]RelayConn) {
	for _, relayConn := range relayConnList {
		conn := *relayConn.conn
		if conn != nil {
			conn.Close()
		}
	}
	if l != nil {
		l.Close()
	}
}

// Accepts new connections on the default HOST_PORT (8080 by default)
// No SBRP protocol messages are returned over TCP at this level as the connection is not established.
// Each attempted connection increments a counter which is then used as the Host connection identifier internally
func waitForHostConnections(l net.Listener, op chan sbrp.Op, hostListChan chan RelayConn) {
	connCount := 0
	for {
		connCount++
		fmt.Println("Listening for host #" + strconv.Itoa(connCount) + " on port " + strconv.Itoa(cfg.HostPort))
		conn, err := l.Accept()
		if err != nil {
			// Use of error string comparison supported in Go 1.9, Go issue #39997
			if strings.Contains(err.Error(), "use of closed") &&
				strings.Contains(err.Error(), "network connection") {
				fmt.Println("Host Listener on port " + strconv.Itoa(cfg.HostPort) + " confirmed closed")
				break
			} else {
				fmt.Errorf("error accepting an inbound connection, waiting for next connection")
			}
		} else {
			fmt.Println("Connection accepted, launching host handler.")
			conn.SetReadDeadline(time.Now().Add(cfg.TimeoutMinutes * time.Minute))
			go handleHostConnection(conn, connCount, hostListChan, op)
		}
	}
}

// Manages:
// Setting up a client listener on a port and exiting gracefully if it cannot create one
// Receiving data from the Host
// Sending data to the Host and exiting gracefully if the host is no longer connected
// defer order: 1) Close client listener, 2) Remove host from conn list, 3) Close connection
func handleHostConnection(conn net.Conn, id int, hostListChan chan RelayConn, opChan chan sbrp.Op) {
	var err error
	defer conn.Close()

	// Struct containing any reference data for this connection
	// A copy of the Host data is kept in the main goroutine
	var host RelayConn
	host.useSbrp = cfg.UseRelayProtocol
	host.id = id + time.Now().Nanosecond()
	host.conn = &conn
	host.op = sbrp.Add

	defer func() {
		host.op = sbrp.Remove
		hostListChan <- host
	}()

	clientConns := make(map[int]RelayConn)
	var l net.Listener
	defer exit(l, clientConns)

	toHostChan := make(chan DataPack, cfg.ReceiveChanQueueSize)
	clientListChan := make(chan RelayConn)
	clientErrChan := make(chan error)
	// Create a client listener on the first available port within our defined port range
	for p := cfg.ClientPortMin; p <= cfg.ClientPortMax; p++ {
		l, err = net.Listen("tcp", ":"+strconv.Itoa(p))
		if err == nil {
			// Fill host information
			host.port = p
			host.name = strconv.Itoa(p)
			// inform the master goroutine that his host is ready
			host.op = sbrp.Add
			hostListChan <- host
			// Start accepting client connections for this host
			go waitForClientConnections(l, p, host.id, host.useSbrp, toHostChan, clientListChan, clientErrChan)
			break
		}
	}

	// If after checking all the ports in our defined range none is available as a new listener,
	// we return an error to the terminal and exit the goroutine function (executing the deferred conn.Close()
	// if USE_RELAY_PROTOCOL is true it will send the error message to the Host connection before closing
	if host.port == 0 {
		err = sbrp.ErrResp(conn, sbrp.CannotListenForClient, host.useSbrp,
			"cannot create host/client pipe, closing connection to host")
		if err != nil {
			fmt.Errorf("Host Error -> " + err.Error())
		}
		return
	} else {
		// Else we inform the user through terminal of the assigned listening port
		// if USE_RELAY_PROTOCOL is true it will send the port message to the Host connection in SBRP format as well
		err = sbrp.HostConnResponse(conn, host.port, host.useSbrp)
		if err != nil {
			fmt.Errorf("Host Error -> " + err.Error())
			sbrp.ErrResp(nil, sbrp.WritingToHost, host.useSbrp,
				"Error Writing to host, closing connection")
			return
		}
	}

	recvChan := make(chan DataPack, 10)
	recvErrorChan := make(chan error)

	go connRead(conn, recvChan, recvErrorChan)

	connId := 0
	// A loop that blocks until it received a message on an assigned channel and reacts to it.
	doneMsg := sbrp.OkState
	for doneMsg == sbrp.OkState {
		select {
		case recvData := <-recvChan:
			fmt.Println(string(recvData.data))
			op, _, _, _ := handleReceivedHostData(conn, host.useSbrp, recvData)
			switch op {
			case sbrp.Quit:
				opChan <- sbrp.Quit
			case sbrp.SendToAllClients:
				err := sendToAllClients(clientConns, recvData)
				if err != nil {
					fmt.Errorf(err.Error())
				}
			default:
				err := sendToAllClients(clientConns, recvData)
				if err != nil {
					fmt.Errorf(err.Error())
				}
			}
		case recvData := <-toHostChan:
			fmt.Println("Data from client #" + strconv.Itoa(recvData.idSource) + "-> " + string(recvData.data))
			conn.Write(recvData.data)

		case rc := <-clientListChan:
			switch rc.op {
			case sbrp.Add:
				connId++
				rc.id = connId
				clientConns[connId] = rc
			case sbrp.Remove:
				delete(clientConns, rc.id)
			}

		case err := <-recvErrorChan:
			fmt.Println("Host Error -> ", err.Error())
			sbrp.ErrResp(conn, sbrp.CannotReceiveDataFromHost, host.useSbrp, err.Error())
			doneMsg = sbrp.CloseState
		case err := <-clientErrChan:
			sbrp.ErrResp(conn, sbrp.ClientClosureError, host.useSbrp, err.Error())
		}
	}
	return
}

func sendToAllClients(clientConns map[int]RelayConn, data DataPack) (err error) {
	for _, c := range clientConns {
		conn := *c.conn
		_, errTmp := conn.Write(data.data)
		if errTmp != nil {
			fmt.Println(c.idParent, c.id, c.port)
			err = errTmp
		}
	}
	return
}

// If any data is received from the Host connection, it is analysed in this function and converted into
// predetermined operations.
func handleReceivedHostData(conn net.Conn, useSbrp bool, d DataPack) (op sbrp.Op, cmd string, data string, err error) {
	prefixLen := len(sbrp.Prefix)
	if d.length >= (prefixLen+sbrp.CmdLength) && string(d.data[:prefixLen]) == sbrp.Prefix {
		op, data, err = sbrp.HandleSbrpRequest(d.data)
		if err != nil {
			err = sbrp.ErrResp(conn, sbrp.BadlyFormattedSbrpMsg, useSbrp, err.Error())
			if err != nil {
				fmt.Errorf("Host Error -> " + err.Error())
				return
			}
		}
	}
	return
}

// A function that waits for client connections on the assigned listener.
// When a client connection is established, it will launch a new goroutine to handle said communication.
func waitForClientConnections(l net.Listener,
	p int,
	idParent int,
	useSbrp bool,
	toHostChan chan DataPack,
	clientListChan chan RelayConn,
	errChan chan error) {
	connCount := 0
	for {
		connCount++
		fmt.Println("Listening for client #" + strconv.Itoa(connCount) + " on port " + strconv.Itoa(p))
		conn, err := l.Accept()
		if err != nil {
			// Use of error string comparison supported in Go 1.9, Go issue #39997
			fmt.Println(err.Error())
			if strings.Contains(err.Error(), "use of closed") &&
				strings.Contains(err.Error(), "network connection") {
				fmt.Println("Client Listener on port " + strconv.Itoa(p) + " confirmed closed")
				break
			} else {
				fmt.Errorf("error on inbound client connection, waiting for next connection")
			}

		} else {
			fmt.Println("Connection accepted, launching host handler.")
			conn.SetReadDeadline(time.Now().Add(cfg.TimeoutMinutes * time.Minute))
			go handleClientConnection(conn, connCount, idParent, useSbrp, toHostChan, clientListChan, errChan)
		}
	}
}

// Manages:
// Setting up a client listener on a port and exiting gracefully if it cannot create one
// Receiving data from the Host
// Sending data to the Host and exiting gracefully if the host is no longer connected
// defer order: 1) Remove host from conn list, 2) Close connection
func handleClientConnection(conn net.Conn,
	id int,
	idParent int,
	useSbrp bool,
	toHostChan chan DataPack,
	clientListChan chan RelayConn,
	errChan chan error) {
	defer conn.Close()

	// Struct containing any reference data for this connection
	// A copy of the Host data is kept in the main goroutine
	var client RelayConn
	client.useSbrp = useSbrp
	client.id = id
	client.idParent = idParent
	client.conn = &conn
	client.op = sbrp.Add

	clientListChan <- client

	defer func() {
		client.op = sbrp.Remove
		clientListChan <- client
	}()

	recvChan := make(chan DataPack, cfg.ReceiveChanQueueSize)
	recvErrorChan := make(chan error)

	go connRead(conn, recvChan, recvErrorChan)

	// A loop that blocks until it received a message on an assigned channel and reacts to it.
	doneMsg := sbrp.OkState
	for doneMsg == sbrp.OkState {
		select {
		case recvData := <-recvChan:
			recvData.idSource = client.id
			toHostChan <- recvData
		case err := <-recvErrorChan:
			fmt.Println("Client Error -> ", err.Error())
			errChan <- err
			doneMsg = sbrp.CloseState
		}
	}
}

// Generic function that handles the traditional blocking socket.read() functionality
func connRead(conn net.Conn, recvChan chan DataPack, errChan chan error) {
	var readBuff = make([]byte, cfg.ReceiveBufferSize)
	for {
		readLen, err := conn.Read(readBuff)
		if err != nil {
			fmt.Errorf("Connection Read Error -> " + err.Error())
			errChan <- err
			break
		}
		if readLen == 0 {
			err = errors.New("can no longer receive data from connection")
			fmt.Errorf(err.Error())
			errChan <- err
			break
		} else {
			var dp DataPack
			dp.data = make([]byte, readLen)
			copy(dp.data, readBuff)
			dp.length = readLen
			fmt.Println("len-> " + strconv.Itoa(dp.length) + " " + string(dp.data))
			recvChan <- dp
		}
	}
}
