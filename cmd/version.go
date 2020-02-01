package cmd

import (
	"fmt"
	"os"

	"github.com/reservoird/reservoird/ver"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version information",
	Run: func(cmd *cobra.Command, args []string) {
		version := ver.NewVersion()
		fmt.Printf("%s\n", version.String())
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
