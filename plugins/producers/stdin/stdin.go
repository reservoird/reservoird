package main

import (
	"bufio"
	"os"
)

type stdin struct {
}

// Produce reads messages from stdin and writes them to a channel
func (stdin *stdin) Produce(channel chan<- []byte) error {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		select {
		case channel <- []byte(line):
		}
	}
}

// Producer for stdin
var Producer stdin
