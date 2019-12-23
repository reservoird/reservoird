package main

type fwd struct {
}

// Config configures consumer
func (o *fwd) Config(cfg string) error {
	return nil
}

// Format reads from src channel and forwards to dst channel
func (o *fwd) Format(src <-chan []byte, dst chan<- []byte) error {
	for {
		select {
		case line := <-src:
			dst <- line
		}
	}
}

// Formatter for fwd
var Formatter fwd
