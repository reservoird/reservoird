package main

import (
	"fmt"
)

type stdout struct {
}

// Consume reads messages from a channel and writes them to stdout
func (stdout *stdout) Consume(channel <-chan []byte) error {
	for {
		select {
		case line := <-channel:
			fmt.Printf("%s", line)
		}
	}
}

// Consumer for stdout
var Consumer stdout
