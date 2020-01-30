package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/reservoird/reservoird/cfg"
	"github.com/reservoird/reservoird/run"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a reservoird config",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if Debug == true {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}
		log.SetFormatter(&log.JSONFormatter{})

		data, err := ioutil.ReadFile(args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		rsv := cfg.Cfg{}
		err = json.Unmarshal(data, &rsv)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		log.Info("=== beg ===")

		reservoirMap, err := run.NewReservoirMap(rsv)
		if err != nil {
			log.Fatalf("error setting up reservoirs: %v\n", err)
		}
		reservoirMap.StartAll()

		server, err := run.NewServer(reservoirMap, Address)
		if err != nil {
			log.Fatalf("error setting up server: %v\n", err)
		}
		server.RunMonitor()

		err = server.Serve()
		if err != nil {
			log.Fatalf("error serving rest interface: %v\n", err)
		}

		reservoirMap.InitStopAll()
		server.StopMonitor()
		reservoirMap.WaitAll()

		log.Info("=== end ===")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
