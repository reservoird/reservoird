package cmd

import (
	"github.com/spf13/cobra"
)

var Debug bool
var rootCmd = &cobra.Command{
	Use:   "reservoird",
	Short: "Reservoird is a pluggable light-weight stream processing framework",
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "debug mode")
}

func Execute() error {
	return rootCmd.Execute()
}
