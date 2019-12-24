package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

type stdout struct {
	Timestamp bool
}

// Config configures consumer
func (o *stdout) Config(cfg string) error {
	// default
	o.Timestamp = false
	if cfg != "" {
		b, err := ioutil.ReadFile(cfg)
		if err != nil {
			return err
		}
		s := stdout{}
		err = json.Unmarshal(b, &s)
		if err != nil {
			return err
		}
		o.Timestamp = s.Timestamp
	}
	return nil
}

// Expel reads messages from a channel and writes them to stdout
func (o *stdout) Expel(channel <-chan []byte) error {
	for {
		line := <-channel

		if o.Timestamp == true {
			fmt.Printf("[stdout %s]: %s", time.Now().Format(time.RFC3339), line)
		} else {
			fmt.Printf("%s", line)
		}
	}
}

// Expeller for stdout
var Expeller stdout
