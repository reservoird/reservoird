package run

// Producer provides interface for producers
type Producer interface {
	Config(string) error
	Produce(chan<- []byte) error
}

// Formatter provides interface for formatters
type Formatter interface {
	Config(string) error
	Format(<-chan []byte, chan<- []byte) error
}

// Consumer provides interface for consumers
type Consumer interface {
	Config(string) error
	Consume(<-chan []byte) error
}
