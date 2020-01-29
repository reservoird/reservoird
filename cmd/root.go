package cmd

import (
	"github.com/spf13/cobra"
)

var Debug bool
var Config string
var rootCmd = &cobra.Command{
	Use:   "reservoird",
	Short: "Reservoird is a pluggable light-weight stream processing framework",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		Config = args[0]
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "debug mode")
}

func Execute() error {
	return rootCmd.Execute()
}
