package run

import (
	"sync"
)

// Ingester provides interface for plugins that ingest data into reservoird
// struct channel and wait group are for graceful shutdown
type Ingester interface {
	Config(string) error
	Ingest(chan<- []byte, <-chan struct{}, *sync.WaitGroup) error
}

// Digester provides interface for plugins that filter/annotate data within reservoird
// struct channel and wait group are for graceful shutdown
type Digester interface {
	Config(string) error
	Digest(<-chan []byte, chan<- []byte, <-chan struct{}, *sync.WaitGroup) error
}

// Expeller provides interface for plugins that expel data outof reservoird
// struct channel and wait group are for graceful shutdown
type Expeller interface {
	Config(string) error
	Expel(<-chan []byte, <-chan struct{}, *sync.WaitGroup) error
}
