package cmd

import (
	"github.com/spf13/cobra"
)

var Address string
var Debug bool
var rootCmd = &cobra.Command{
	Use:   "reservoird",
	Short: "Reservoird is a pluggable light-weight stream processing framework",
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&Address, "address", "a", ":5514", "host and port")
	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "debug mode")
}

func Execute() error {
	return rootCmd.Execute()
}
