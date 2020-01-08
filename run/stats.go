package run

import (
	"runtime"
	"runtime/debug"
)

// RuntimeStats provide go runtime stats
type RuntimeStats struct {
	CPUs       int               `json:"cpus"`
	Goroutines int               `json:"goroutines"`
	Goversion  string            `json:"goversion"`
	BuildInfo  *debug.BuildInfo  `json:"modules"`
	MemStats   *runtime.MemStats `json:"mem"`
	GCStats    *debug.GCStats    `json:"gc"`
}

// FlowStats provide reservoir flow stats
type FlowStats map[string][]string
