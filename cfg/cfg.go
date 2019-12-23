package cfg

// ProducerItemCfg contains the configuration for a producer
type ProducerItemCfg struct {
	Location    string `json:"location"`
	ConfigFile  string `json:"configFile"`
	ChannelSize int    `json:"channelSize"`
}

// FormatterItemCfg contains the configuration for a formatter
type FormatterItemCfg struct {
	Location    string `json:"location"`
	ConfigFile  string `json:"configFile"`
	ChannelSize int    `json:"channelSize"`
}

// ConsumerItemCfg contains the configuration for a consumer
type ConsumerItemCfg struct {
	Location   string `json:"location"`
	ConfigFile string `json:"configFile"`
}

// ReservoirCfg contains the configuration for the flow
type ReservoirCfg struct {
	Producer   ProducerItemCfg    `json:"producer"`
	Formatters []FormatterItemCfg `json:"formatters"`
	Consumer   ConsumerItemCfg    `json:"consumer"`
}

// Cfg configures system
type Cfg struct {
	Reservoirs []ReservoirCfg `json:"reservoirs"`
}
