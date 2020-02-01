package sta

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

// FlowStats provides reservoir flow stats
type FlowStats map[string][]string

// ReservoirStats provides reservoir stats
type ReservoirStats map[string][]interface{}

// Version
type Version struct {
	GitVersion string `json:"gitVersion"`
	GitHash    string `json:"gitHash"`
	GoVersion  string `json:"goVersion"`
	ICDPath    string `json:"icdPath"`
	ICDVersion string `json:"icdVersion"`
}
