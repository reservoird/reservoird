package cfg

// QueueItemCfg contains the configuration for a queue
type QueueItemCfg struct {
	Location string `json:"location"`
	Config   string `json:"config"`
}

// IngesterItemCfg contains the configuration for an ingester
type IngesterItemCfg struct {
	Location  string            `json:"location"`
	Config    string            `json:"config"`
	QueueItem QueueItemCfg      `json:"queue"`
	Digesters []DigesterItemCfg `json:"digesters"`
}

// DigesterItemCfg contains the configuration for a digester
type DigesterItemCfg struct {
	Location  string       `json:"location"`
	Config    string       `json:"config"`
	QueueItem QueueItemCfg `json:"queue"`
}

// ExpellerItemCfg contains the configuration for an expeller
type ExpellerItemCfg struct {
	Location      string            `json:"location"`
	Config        string            `json:"config"`
	IngesterItems []IngesterItemCfg `json:"ingesters"`
}

// ReservoirCfg contains the configuration for the flow
type ReservoirCfg struct {
	Name         string          `json:"name"`
	ExpellerItem ExpellerItemCfg `json:"expeller"`
}

// Cfg configures system
type Cfg struct {
	Reservoirs []ReservoirCfg `json:"reservoirs"`
}
