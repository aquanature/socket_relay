package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

type Cfg struct {
	HostPort             int           `json:"host_post"`
	TimeoutMinutes       time.Duration `json:"timeout_minutes"`
	ClientPortMin        int           `json:"client_port_min"`
	ClientPortMax        int           `json:"client_port_max"`
	UseRelayProtocol     bool          `json:"use_relay_protocol"`
	ReceiveBufferSize    int           `json:"receive_buffer_size"`
	ReceiveChanQueueSize int           `json:"receive_queue_size"`
}

func (c *Cfg) Init() {
	f, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Errorf("cannot find config.json, using coded defaults")
		c.SetDefaults()
	} else {
		json.Unmarshal(f, &c)
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
