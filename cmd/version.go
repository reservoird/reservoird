package cmd

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

const (
	// GitVersion is the version based off git describe
	GitVersion string = "v0.0.0"
	// GitHash is the git hash
	GitHash string = "n/a"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", GetVersion())
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

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

	return fmt.Sprintf("%s (%s) [%s] [%s %s]\n",
		GitVersion,
		GitHash,
		runtime.Version(),
		icdpath,
		icdversion,
	)
}
