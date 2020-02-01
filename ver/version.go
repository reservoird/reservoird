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
)

func GetVersion() string {
	icdpath := "github.com/reservoird/icd"
	icdversion := "unknown"
	buildinfo, ok := debug.ReadBuildInfo()
	if ok == true {
		for i := range buildinfo.Deps {
			if buildinfo.Deps[i].Path == icdpath {
				icdversion = buildinfo.Deps[i].Version
			}
		}
	}

	return fmt.Sprintf("%s (%s) [%s] [%s %s]",
		GitVersion,
		GitHash,
		runtime.Version(),
		icdpath,
		icdversion,
	)
}
