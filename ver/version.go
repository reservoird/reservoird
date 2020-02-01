package ver

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

const (
	// GitVersion is the version based off git describe
	GitVersion = "v0.0.0"
	// GitHash is the git hash
	GitHash string = "n/a"
	// ICDPath is the path of the icd library
	ICDPath string = "github.com/reservoird/icd"
)

// Version
type Version struct {
	GitVersion string `json:"gitVersion"`
	GitHash    string `json:"gitHash"`
	GoVersion  string `json:"goVersion"`
	ICDPath    string `json:"icdPath"`
	ICDVersion string `json:"icdVersion"`
}

func NewVersion() *Version {
	o := &Version{
		GitVersion: GitVersion,
		GitHash:    GitHash,
		GoVersion:  runtime.Version(),
		ICDPath:    ICDPath,
		ICDVersion: "unknown",
	}

	buildinfo, ok := debug.ReadBuildInfo()
	if ok == true {
		for i := range buildinfo.Deps {
			if buildinfo.Deps[i].Path == ICDPath {
				o.ICDVersion = buildinfo.Deps[i].Version
			}
		}
	}

	return o
}

func (o *Version) String() string {
	return fmt.Sprintf("%s (%s) [%s] [%s %s]",
		o.GitVersion,
		o.GitHash,
		o.GoVersion,
		o.ICDPath,
		o.ICDVersion,
	)
}
