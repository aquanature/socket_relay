package main

import (
	"flag"
	"fmt"
	"github.com/aquanature/socket_relay/relay"
)

// Main manages:
// - The global channels
// - The master list of currently managed hosts
// - Manage any exit channels and signals
// - Manage the graceful exiting of all goroutines, especially the closing of all OS/System connections.
func main() {
	isSimpleServerAdapter := *flag.Bool("s", false, "launches the app as a simple relay server adapter")
	isMultiplexingServerAdapter := *flag.Bool("m", false, "launches the app as a multiplexing relay server adapter")
	isEchoHost := *flag.Bool("e", false, "launches the app as an echo server")
	isRelay := *flag.Bool("r", false, "launches the app as a relay")
	desiredPort := *flag.Int("port", 0, "port the relay or server adpater listens on for new connections")

	if isSimpleServerAdapter {

	} else if isMultiplexingServerAdapter {

	} else if isEchoHost {

	} else {
		relay.MainLoop(desiredPort)
	}

	fmt.Print(isSimpleServerAdapter, isMultiplexingServerAdapter, isRelay, isEchoHost)
}