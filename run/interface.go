package run

import (
	"sync"
)

// Queue provides interface for queue plugins
type Queue interface {
	Config(string) error
	Push(interface{}) error
	Pop() (interface{}, error)
	Close() error
}

// Ingester provides interface for plugins that ingest (push) data into reservoird
// struct channel and wait group are for graceful shutdown of ingester plugin
type Ingester interface {
	Config(string) error
	Ingest(Queue, <-chan struct{}, *sync.WaitGroup) error
}

// Digester provides interface for plugins that filter/annotate (push/pop) data within reservoird
// struct channel and wait group are for graceful shutdown of digester plugin
type Digester interface {
	Config(string) error
	Digest(Queue, Queue, <-chan struct{}, *sync.WaitGroup) error
}

// Expeller provides interface for plugins that expel (pop) data outof reservoird
// struct channel and wait group are for graceful shutdown of expeller plugin
type Expeller interface {
	Config(string) error
	Expel(Queue, <-chan struct{}, *sync.WaitGroup) error
}
