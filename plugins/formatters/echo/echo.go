package main

type echo struct {
}

// Format reads from src channel and echos that to dst channel
func (echo *echo) Format(src <-chan []byte, dst chan<- []byte) error {
	for {
		select {
		case line := <-src:
			dst <- line
		}
	}
}

// Formatter for echo
var Formatter echo
