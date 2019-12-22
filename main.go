package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/reservoird/reservoird/cfg"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	// GitVersion is the version based off git describe
	GitVersion string = "v0.0.0"
	// GitHash is the git hash
	GitHash string = "n/a"
	// REST address kind
	REST string = "rst"
	// TCP address kind
	TCP string = "tcp"
	// UDP address kind
	UDP string = "udp"
)

func handler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
}

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
		fmt.Printf("error reading %s: %v", flag.Args()[0], err)
		os.Exit(1)
	}

	rsv := cfg.Cfg{}
	err = json.Unmarshal(data, &rsv)
	if err != nil {
		fmt.Printf("error parsing %s: %v", flag.Args()[0], err)
		os.Exit(1)
	}

	switch rsv.Rsv.Kind {
	case REST:
		http.HandleFunc("/", handler)
		err = http.ListenAndServe(rsv.Rsv.Address, nil)
		if err != nil {
			fmt.Printf("%v", err)
		}
	case TCP:
		fmt.Printf("todo kind %s\n", rsv.Rsv.Kind)
	case UDP:
		fmt.Printf("todo kind %s\n", rsv.Rsv.Kind)
	default:
		fmt.Printf("unknown kind %s\n", rsv.Rsv.Kind)
	}

}
