package sbrp

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	Prefix    = "SBRP 1.0 "
	CmdLength = 10
)

// Protocol error codes
type Error int

const (
	OkState Error = 0 + iota
	CloseState
)

type Op int

const (
	Unknown Op = -1 + iota
	Add
	Remove
	Rename
	ListConns
	UseProtocol
	Loop
	SendToAllClients
	Quit
)

// Protocol error codes
const (
	Ok Error = 0 + iota
	CannotListenForHost
	CannotListenForClient
	WritingToHost
	BadlyFormattedSbrpMsg
	CannotReceiveDataFromHost
	ClientClosureError
)

func (errorCode Error) ToStr() (errorName string) {
	// temporary code until the list is finalized for prototype
	errorName = "TODO: will create name from error code " + strconv.Itoa(int(errorCode))
	return
}

func (op Op) ToStr() (opName string) {
	// temporary code until the list is finalized for prototype
	opName = "Operation # " + strconv.Itoa(int(op))
	return
}

func HandleSbrpRequest(rawData []byte) (op Op, msg string, err error) {
	cmd := string(rawData[9:19])
	if len(rawData) > 20 {
		msg = string(rawData[20:])
	}
	fmt.Println(cmd + "-" + msg)
	switch cmd {
	case "RENAME_CON":
		op = Rename
	case "QUIT_RELAY":
		op = Quit
	case "LIST_CONNS":
		op = ListConns
	case "SWITCHSBRP":
		op = UseProtocol
		if strings.Contains(msg, "OFF") || strings.Contains(msg, "FALSE") {
			msg = "FALSE"
		} else {
			msg = "TRUE"
		}
	default:
		op = Unknown
	}
	return
}

func ErrResp(conn net.Conn, errCode Error, useSbrp bool, errMsg string) (err error) {
	errorResp := Prefix + "ERROR_MESG " + errCode.ToStr() + " " + errMsg + "\r\n"
	fmt.Errorf(errorResp)
	if useSbrp && conn != nil {
		_, err = conn.Write([]byte(errorResp))
	}
	return err
}

func HostConnResponse(conn net.Conn, p int, useSbrp bool) (err error) {
	resp := Prefix + "RELAY_PORT " + strconv.Itoa(p) + "\r\n"
	fmt.Println(resp)
	if useSbrp && conn != nil {
		_, err = conn.Write([]byte(resp))
	}
	return err
}
