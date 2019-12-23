package main

import (
	"fmt"
)

type echo struct {
}

// Format reads from src channel echos to screen and forwards to dst channel
func (echo *echo) Format(src <-chan []byte, dst chan<- []byte) error {
	for {
		select {
		case line := <-src:
			fmt.Printf("%s", line)
			dst <- line
		}
	}
}

// Formatter for echo
var Formatter echo
