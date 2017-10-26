package sbrp

import (
	"fmt"
	"net"
	"strconv"
)

const (
	Prefix = "SBRP 1.0 "
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
	Unknown Op = -1
	Modify Op = 0
	Add    Op = 0 + iota
	Remove
	Rename
	ListConns
	Quit
)

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
	default:
		op = Unknown
	}
	return
}

func ErrResp(conn net.Conn, errCode int, useSbrp bool, errMsg string) (err error) {
	errorResp := Prefix + "ERROR_MESG " + strconv.Itoa(errCode) + " " + errMsg + "\r\n"
	fmt.Errorf(errorResp)
	if useSbrp && conn != nil {
		_, err = conn.Write([]byte(errMsg))
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
