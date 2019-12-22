package cfg

// RsvCfg configures reservoird process
type RsvCfg struct {
	Kind    string `json:"kind"`
	Address string `json:"address"`
}

// Cfg configures system
type Cfg struct {
	Rsv RsvCfg `json:"reservoird"`
}
