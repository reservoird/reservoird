package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/reservoird/reservoird/cfg"
	"github.com/reservoird/reservoird/run"
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
	flag.BoolVar(&help, "help", false, "print help")
	flag.BoolVar(&version, "version", false, "print version")
	flag.Parse()

	if version == true {
		fmt.Printf("%s (%s)\n", GitVersion, GitHash)
		os.Exit(0)
	}

	if help == true {
		flag.Usage()
		os.Exit(0)
	}

	if len(flag.Args()) == 0 {
		fmt.Printf("configuration filename required\n")
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(flag.Args()[0])
	if err != nil {
		fmt.Printf("error reading %s: %v\n", flag.Args()[0], err)
		os.Exit(1)
	}

	rsv := cfg.Cfg{}
	err = json.Unmarshal(data, &rsv)
	if err != nil {
		fmt.Printf("error parsing %s: %v\n", flag.Args()[0], err)
		os.Exit(1)
	}

	reservoirs, err := run.NewReservoirs(rsv)
	if err != nil {
		fmt.Printf("error setting up reservoirs: %v\n", err)
		os.Exit(1)
	}
	err = run.Run(reservoirs)
	if err != nil {
		fmt.Printf("error running: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("done.\n")
}
