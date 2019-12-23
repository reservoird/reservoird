package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type stdout struct {
	Prefix string
	Repeat int
}

// Config configures consumer
func (o *stdout) Config(cfg string) error {
	// default
	o.Prefix = "stdout"
	o.Repeat = 1
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
		o.Prefix = s.Prefix
		o.Repeat = s.Repeat
	}
	return nil
}

// Consume reads messages from a channel and writes them to stdout
func (o *stdout) Consume(channel <-chan []byte) error {
	for {
		select {
		case line := <-channel:
			for i := 0; i < o.Repeat; i++ {
				fmt.Printf("%s: %s", o.Prefix, line)
			}
		}
	}
}

// Consumer for stdout
var Consumer stdout
