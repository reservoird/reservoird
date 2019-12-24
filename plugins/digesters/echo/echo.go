package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

type echo struct {
	Timestamp bool
}

// Config configures consumer
func (o *echo) Config(cfg string) error {
	// default
	o.Timestamp = false
	if cfg != "" {
		b, err := ioutil.ReadFile(cfg)
		if err != nil {
			return err
		}
		s := echo{}
		err = json.Unmarshal(b, &s)
		if err != nil {
			return err
		}
		o.Timestamp = s.Timestamp
	}
	return nil
}

// Format reads from src channel echos to screen and forwards to dst channel
func (o *echo) Format(src <-chan []byte, dst chan<- []byte) error {
	for {
		select {
		case line := <-src:
			if o.Timestamp == true {
				dst <- []byte(fmt.Sprintf("%s: %s", time.Now().Format(time.RFC3339), line))
			}
			dst <- line
		}
	}
}

// Formatter for echo
var Formatter echo
