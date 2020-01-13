package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/reservoird/reservoird/cfg"
	"github.com/reservoird/reservoird/run"
	log "github.com/sirupsen/logrus"
)

const (
	// GitVersion is the version based off git describe
	GitVersion string = "v0.0.0"
	// GitHash is the git hash
	GitHash string = "n/a"
)

func main() {
	var help bool
	var version bool
	var debugging bool
	flag.BoolVar(&help, "help", false, "print help")
	flag.BoolVar(&version, "version", false, "print version")
	flag.BoolVar(&debugging, "debug", false, "turn on debugging")

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("reservoird <config>\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if version == true {
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

		fmt.Printf("%s (%s) [%s] [%s %s]\n",
			GitVersion,
			GitHash,
			runtime.Version(),
			icdpath,
			icdversion,
		)
		os.Exit(0)
	}

	if help == true {
		flag.Usage()
		os.Exit(0)
	}

	if debugging == true {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.SetFormatter(&log.JSONFormatter{})

	if len(flag.Args()) == 0 {
		log.Fatalf("error configuration filename required\n")
	}

	log.Info("=== beg ===")

	data, err := ioutil.ReadFile(flag.Args()[0])
	if err != nil {
		log.Fatalf("error reading configuration file (%s): %v", flag.Args()[0], err)
	}

	rsv := cfg.Cfg{}
	err = json.Unmarshal(data, &rsv)
	if err != nil {
		log.Fatalf("error unmarshalling configuration file (%s): %v\n", flag.Args()[0], err)
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
