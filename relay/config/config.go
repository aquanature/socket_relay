package config

import "time"

type Cfg struct {
	HostPort             int
	TimeoutMinutes       time.Duration
	ClientPortMin        int
	ClientPortMax        int
	UseRelayProtocol     bool
	ReceiveBufferSize    int
	ReceiveChanQueueSize int
}

func (c *Cfg) Init() (ok bool) {
	ok = false //this is a load from file failure simulation
	if !ok {
		c.SetDefaults()
	}
	return
}

func (c *Cfg) SetDefaults() {
	c.HostPort = 8080
	c.TimeoutMinutes = 5
	c.ClientPortMin = 8081
	c.ClientPortMax = 23000
	c.UseRelayProtocol = true
	c.ReceiveBufferSize = 512
	c.ReceiveChanQueueSize = 10
}
