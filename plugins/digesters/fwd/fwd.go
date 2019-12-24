package main

type fwd struct {
}

// Config configures consumer
func (o *fwd) Config(cfg string) error {
	return nil
}

// Digest reads from src channel and forwards to dst channel
func (o *fwd) Digest(src <-chan []byte, dst chan<- []byte) error {
	for {
		select {
		case line := <-src:
			dst <- line
		}
	}
}

// Digester for fwd
var Digester fwd
