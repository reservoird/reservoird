package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type stdin struct {
	Timestamp bool
}

// Config configures consumer
func (o *stdin) Config(cfg string) error {
	// default
	o.Timestamp = false
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
		o.Timestamp = s.Timestamp
	}
	return nil
}

// Ingest reads messages from stdin and writes them to a channel
func (o *stdin) Ingest(channel chan<- []byte) error {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if o.Timestamp == true {
			channel <- []byte(fmt.Sprintf("[stdin %s]: %s", time.Now().Format(time.RFC3339), line))
		} else {
			channel <- []byte(line)
		}
	}
}

// Ingester for stdin
var Ingester stdin
