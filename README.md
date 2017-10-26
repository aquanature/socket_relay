# Simple Multi-Protocol TCP Relay
by Stephane Bedard, sbedard.code@gmail.com

## Introduction

The Relay (as the Simple Multi-Protocol TCP Relay is hence forth referred to) allows two TCP networked applications to 
communicate with each other when a direct communication is not possible, for example if the two applications are hosted
on different NAT sub-networks. 

The Relay supports 3 basic modes of operation: 
- One to One Mode
- One to Many Mode
- Clients/Server with simple Server Adapter
- Client(s)/Server(s) mode with Multiplexing Server Adapter

#
## Supported OS
- Compiled: Windows 10 64 bit 
- Compiled and provided binary: Ubuntu 16.04.3 LTS (should work with most flavors of Linux)


#
## Glossary

Client: A connection that connects to the port assigned to a Host for the purpose of communicating with the Host.

Host: A networked connection that simulates listening for other connections. Any commands to the relay must come from
this connection. This is also the connection that a server adapter application would connect to.

Server Adapter: An application with an outbound connection to the Relay that can forward requests to outside servers
and multiplex the response back to the Relay for the client that made the original request. In theory the
Server Adapter allows communication between a traditional server and a client with no modifications. 
More details about the Server Adapter specifications can be found in the designated section of this document.

SBRP: The Relay protocol that the Relay uses to send messages to the Host and that the Host can use to send commands 
to the Relay (Quit, Rename Host, List all Hosts, etc.). More details on the SBRP protocol under the designated section
of this document.

#
## One to One Mode

One to One mode allows for traditional peer to peer communication between two applications. What one application
transmits, the other receives and vice versa.

This is the default behavior when a single client connects to the assigned port of a single host. An example use
could be for example for a Telnet chat.

#
## One to Many Mode

One to Many mode allows a Host connection to broadcast to all connected clients and for all the Clients to message the 
Host connection. The clients do not see each others messages.

This is the default behavior when a multiple clients connects to the assigned port of a single host. An example use
could be the host logging all data from multiple clients, with the capability of sending commands back to the clients.

#
## One to One Client/Server with simple Server Adapter

Allows client/server TCP communication.

The single Host in this scenario is the Server Adapter application connected to the Relay. 
The Client is a traditional client such as a web browser or SSH client. The Server Adapter
would be configured to be able to forward requests to a single IP address and port per connection to the Relay.

- The Relay forwards the unaltered Client request to the Server Adapter.
- If the message is not an SBRP message, the Server Adapter makes a request to a redetermined IP.
- Upon a response, the Server Adapter will forward to the Relay the unaltered response.
- If more than one Client is connected to the Host, all clients will receive the response, whether they made the request
or not.

More details can be found in the Server Adapter specifications.

#
## Client(s)/Server(s) mode with Multiplexing Server Adapter

Allows client/server TCP communication.

The Hosts in this scenario are multiple Server Adapter application that have multiplexing capabilities,
connected to the Relay. The Clients are traditional clients such as web browsers or SSH clients. The Server Adapter
would be configured to be able to forward requests to a single IP address and port per connection to the Relay. 

- It works by the Server Adapter sending a SBRP command to the Relay.
- The Relay will then prepend all Client->Host communication with a client identifier.
- The Server Adapter will strip the identifier before making any outbound requests to a redetermined IP.
- Upon a response, the Server Adapter will once again prepend the client identifier.
- The Relay will route the data to the correct Client, stripping the client identifier before sending.
- The Server Adapter would also not forward any SBRP protocol messages originating from the Relay to the server.

More details can be found in the Server Adapter specifications.

#
## Supported Protocols

These protocols have been tested (or will be tested before release).

- Raw TCP communication peer to peer
- Telnet
- SSH
- HTTP
- Modbus
- SMTP
- FTP

#
## Unsupported Protocols
- HTTPS
- TLS
- UDP

#
## Limitations
- TCP packets containing more than 512 bytes of data in their data field will be split into multiple packets.
- The Relay has no support for partial packet transmission, it will forward what it receives, even if it is corrupted.
This is an unlikely scenario on Windows or Linux where there is a well implemented TCP stack.


#
## Compiling the Relay from GitHub source

TODO

#
## Running Automated Tests

TODO

#
## Starting the Relay

TODO Windows
TODO Linux

#
## Starting the Relay for client(s)/server(s) communication

TODO Simple Server Adapter Application
TODO Multiplexing Server Adapter Application

#
##Using the provided test Server Adapter Application

TODO Windows
TODO Linux

#
## Server Adapter Specification
TODO more details
#### Simple:


#### Multiplexing:
In short, the first 8 bytes of the TCP message, if the Server Adapter is active, are used as a client identifier.
These identifying bytes are only "on the wire" between the Relay and the Server Adapter.

For simplification of testing purposes, the identifier is an 8 character numeral ranging from 00,000,000 to 99,999,999 
(without the commas of course).


#
## SBRP Protocol

The SBRP protocol is designed to use the same syntax both for incoming and outgoing messages.

The SBRP message starts with the identifier "SBRP 1.0 " followed by a 10 character command identifier and 
a trailing space. Anything after the trailing space is considered the message body. In other words, the 21st character 
until the last character is the message body. 

The message body only supports alphanumeric characters from ASCII[32] to ASCII[125]

#
## Host SBRP Messages to the Relay
```
SBRP 1.0 LIST_CONNS HOSTS
SBRP 1.0 LIST_CONNS CLIENTS
SBRP 1.0 LIST_CONNS ALL
SBRP 1.0 LIST_CONNS ME
```
- Returns an unformatted list of connections to the Relay and the ports used
- The ME option returns the information about the connected host the message comes from, as well
as information about any Clients connected to that Host's assigned port.


```
SBRP 1.0 RENAME_CON My_New_Name
```

- Gives a human readable name to the Host connection, useful for identifying a specific host in the list

```
SBRP 1.0 RELAY_PORT GET
```

- Returns the relay port used by the host in this format "SBRP 1.0 RELAY_PORT 8081"

```
SBRP 1.0 QUIT_RELAY
```

#
## Relay SBRP Messages to the Host 

If the const USE_RELAY_PROTOCOL is set to true, messages will be sent to the Host and the console.

If the const USE_RELAY_PROTOCOL is set to false, SBRP messages and errors will only be sent to the console.

```
SBRP 1.0 RELAY_PORT 8081
```

- The first message sent to the Host after connection. As well as the return format of the RELAY_PORT GET command

```
SBRP 1.0 ERROR_MSG Some_Error_String_Goes_Here
```

- Informs the Host of a Relay generated error.



#
## Configuration settings

TODO

#
## Future Improvements

- If required, a future version of the Relay and Server Adapter could use something more akin to NAT, but research would
be required to see if this is possible entirely from the application layer in Go.
- Not only having a minimum and maximum port number assignment for Clients in the configuration but also
having a blacklist of ports that the Relay should not use even within said range.
- Moving all configurations from global const variables in the main.go file to a config.cfg file.
