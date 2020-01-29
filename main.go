package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/reservoird/reservoird/cfg"
	"github.com/reservoird/reservoird/cmd"
	"github.com/reservoird/reservoird/run"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}

	if cmd.Debug == true {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.SetFormatter(&log.JSONFormatter{})

	log.Info("=== beg ===")

	data, err := ioutil.ReadFile(cmd.Config)
	if err != nil {
		log.Fatalf("%v", err)
	}

	rsv := cfg.Cfg{}
	err = json.Unmarshal(data, &rsv)
	if err != nil {
		log.Fatalf("error unmarshalling configuration file (%s): %v\n", cmd.Config, err)
	}

	reservoirMap, err := run.NewReservoirMap(rsv)
	if err != nil {
		log.Fatalf("error setting up reservoirs: %v\n", err)
	}
	reservoirMap.StartAll()

	server, err := run.NewServer(reservoirMap)
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
}
