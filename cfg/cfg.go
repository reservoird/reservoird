package cfg

// IngesterItemCfg contains the configuration for an ingester
type IngesterItemCfg struct {
	Location    string `json:"location"`
	ConfigFile  string `json:"configFile"`
	ChannelSize int    `json:"channelSize"`
}

// DigesterItemCfg contains the configuration for a digester
type DigesterItemCfg struct {
	Location    string `json:"location"`
	ConfigFile  string `json:"configFile"`
	ChannelSize int    `json:"channelSize"`
}

// ExpellerItemCfg contains the configuration for an expeller
type ExpellerItemCfg struct {
	Location   string `json:"location"`
	ConfigFile string `json:"configFile"`
}

// ReservoirCfg contains the configuration for the flow
type ReservoirCfg struct {
	Ingester  IngesterItemCfg   `json:"ingester"`
	Digesters []DigesterItemCfg `json:"digesters"`
	Expeller  ExpellerItemCfg   `json:"expeller"`
}

// Cfg configures system
type Cfg struct {
	Reservoirs []ReservoirCfg `json:"reservoirs"`
}
