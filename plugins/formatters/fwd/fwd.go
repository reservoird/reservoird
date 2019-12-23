package main

type fwd struct {
}

// Format reads from src channel and forwards to dst channel
func (fwd *fwd) Format(src <-chan []byte, dst chan<- []byte) error {
	for {
		select {
		case line := <-src:
			dst <- line
		}
	}
}

// Formatter for fwd
var Formatter fwd
