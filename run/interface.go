package run

// Producer provides interface for producers
type Producer interface {
	Produce(chan<- []byte) error
}

// Formatter provides interface for formatters
type Formatter interface {
	Format(<-chan []byte, chan<- []byte) error
}

// Consumer provides interface for consumers
type Consumer interface {
	Consume(<-chan []byte) error
}
