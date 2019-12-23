package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type stdin struct {
	Prefix string
	Echo   bool
}

// Config configures consumer
func (o *stdin) Config(cfg string) error {
	// default
	o.Prefix = "echo"
	o.Echo = false
	if cfg != "" {
		b, err := ioutil.ReadFile(cfg)
		if err != nil {
			return err
		}
		s := stdin{}
		err = json.Unmarshal(b, &s)
		if err != nil {
			return err
		}
		o.Prefix = s.Prefix
		o.Echo = s.Echo
	}
	return nil
}

// Produce reads messages from stdin and writes them to a channel
func (o *stdin) Produce(channel chan<- []byte) error {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if o.Echo == true {
			fmt.Printf("%s: %s", o.Prefix, line)
		}
		select {
		case channel <- []byte(line):
		}
	}
}

// Producer for stdin
var Producer stdin
