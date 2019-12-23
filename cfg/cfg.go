package cfg

// ReservoirCfg contains the configuration for the flow
type ReservoirCfg struct {
	ChannelSize int      `json:"channelSize"`
	Producer    string   `json:"producer"`
	Formatter   []string `json:"formatter"`
	Consumer    string   `json:"consumer"`
}

// Cfg configures system
type Cfg struct {
	Reservoir []ReservoirCfg `json:"reservoir"`
}
