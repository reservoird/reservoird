package cfg

// QueueItemCfg contains the configuration for a queue
type QueueItemCfg struct {
	Location   string `json:"location"`
	ConfigFile string `json:"configFile"`
}

// IngesterItemCfg contains the configuration for an ingester
type IngesterItemCfg struct {
	Location   string            `json:"location"`
	ConfigFile string            `json:"configFile"`
	QueueItem  QueueItemCfg      `json:"queue"`
	Digesters  []DigesterItemCfg `json:"digesters"`
}

// DigesterItemCfg contains the configuration for a digester
type DigesterItemCfg struct {
	Location   string       `json:"location"`
	ConfigFile string       `json:"configFile"`
	QueueItem  QueueItemCfg `json:"queue"`
}

// ExpellerItemCfg contains the configuration for an expeller
type ExpellerItemCfg struct {
	Location      string            `json:"location"`
	ConfigFile    string            `json:"configFile"`
	IngesterItems []IngesterItemCfg `json:"ingesters"`
}

// ReservoirCfg contains the configuration for the flow
type ReservoirCfg struct {
	ExpellerItem ExpellerItemCfg `json:"expeller"`
}

// Cfg configures system
type Cfg struct {
	Reservoirs []ReservoirCfg `json:"reservoirs"`
}
