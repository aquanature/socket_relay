package error

import "strconv"

// Protocol error codes
const (
	Ok = 0 + iota
	CannotListenForHost
	CannotListenForClient
	WritingToHost
	BadlyFormattedSbrpMsg
	CannotReceiveDataFromHost
)

func ToStr(errorCode int) (errorName string) {
	// temporary code until the list is finalized for prototype
	errorName = "ErrorToStr() will create name from error code " + strconv.Itoa(errorCode)
	return
}
