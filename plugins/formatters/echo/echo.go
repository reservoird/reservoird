package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type echo struct {
	Prefix string
	Repeat int
}

// Config configures consumer
func (o *echo) Config(cfg string) error {
	// default
	o.Prefix = "echo"
	o.Repeat = 1
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
		o.Prefix = s.Prefix
		o.Repeat = s.Repeat
	}
	return nil
}

// Format reads from src channel echos to screen and forwards to dst channel
func (o *echo) Format(src <-chan []byte, dst chan<- []byte) error {
	for {
		select {
		case line := <-src:
			for i := 0; i < o.Repeat; i++ {
				fmt.Printf("%s: %s", o.Prefix, line)
			}
			dst <- line
		}
	}
}

// Formatter for echo
var Formatter echo
