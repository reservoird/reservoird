package run

// Ingester provides interface for plugins that ingest data into reservoird
type Ingester interface {
	Config(string) error
	Ingest(chan<- []byte) error
}

// Digester provides interface for plugins that filter/annotate data within reservoird
type Digester interface {
	Config(string) error
	Digest(<-chan []byte, chan<- []byte) error
}

// Expeller provides interface for plugins that expel data outof reservoird
type Expeller interface {
	Config(string) error
	Expel(<-chan []byte) error
}
