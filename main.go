package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"plugin"

	"github.com/reservoird/reservoird/cfg"
)

const (
	// GitVersion is the version based off git describe
	GitVersion string = "v0.0.0"
	// GitHash is the git hash
	GitHash string = "n/a"
)

// Producer provides interface for producers
type Producer interface {
	Produce(chan<- []byte) error
}

// Formatter provides interface for formatters
type Formatter interface {
	Format(<-chan []byte, chan<- []byte) error
}

// Consumer provides interface for consumers
type Consumer interface {
	Consume(<-chan []byte) error
}

// Reservoir is the structure of the reservoir flow
type Reservoir struct {
	Producer  Producer
	Formatter []Formatter
	Consumer  Consumer
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
		fmt.Printf("error reading %s: %v\n", flag.Args()[0], err)
		os.Exit(1)
	}

	rsv := cfg.Cfg{}
	err = json.Unmarshal(data, &rsv)
	if err != nil {
		fmt.Printf("error parsing %s: %v\n", flag.Args()[0], err)
		os.Exit(1)
	}

	reservoirs := make([]Reservoir, 0)
	for r := range rsv.Reservoir {
		producer, err := plugin.Open(rsv.Reservoir[r].Producer)
		if err != nil {
			fmt.Printf("error loading plugin %s: %v\n", rsv.Reservoir[r].Producer, err)
			os.Exit(1)
		}
		symbolProducer, err := producer.Lookup("Producer")
		if err != nil {
			fmt.Printf("error finding Producer symbol %s: %v\n", rsv.Reservoir[r].Producer, err)
			os.Exit(1)
		}
		prod, ok := symbolProducer.(Producer)
		if ok == false {
			fmt.Printf("error interface not found %s\n", rsv.Reservoir[r].Producer)
			os.Exit(1)
		}

		forms := make([]Formatter, 0)
		for f := range rsv.Reservoir[r].Formatter {
			formatter, err := plugin.Open(rsv.Reservoir[r].Formatter[f])
			if err != nil {
				fmt.Printf("error loading plugin %s: %v\n", rsv.Reservoir[r].Formatter[f], err)
				os.Exit(1)
			}
			symbolFormatter, err := formatter.Lookup("Formatter")
			if err != nil {
				fmt.Printf("error finding Formatter symbol %s: %v\n", rsv.Reservoir[r].Formatter[f], err)
				os.Exit(1)
			}
			form, ok := symbolFormatter.(Formatter)
			if ok == false {
				fmt.Printf("error interface not found %s\n", rsv.Reservoir[r].Formatter[f])
				os.Exit(1)
			}
			forms = append(forms, form)
		}

		consumer, err := plugin.Open(rsv.Reservoir[r].Consumer)
		if err != nil {
			fmt.Printf("error loading plugin %s: %v\n", rsv.Reservoir[r].Producer, err)
			os.Exit(1)
		}
		symbolConsumer, err := consumer.Lookup("Consumer")
		if err != nil {
			fmt.Printf("error finding Consumer symbol %s: %v\n", rsv.Reservoir[r].Producer, err)
			os.Exit(1)
		}
		cons, ok := symbolConsumer.(Consumer)
		if ok == false {
			fmt.Printf("error interface not found %s\n", rsv.Reservoir[r].Producer)
			os.Exit(1)
		}

		reservoirs = append(reservoirs, Reservoir{prod, forms, cons})
	}

	for r := range reservoirs {
		prod := make(chan []byte, 1)
		go reservoirs[r].Producer.Produce(prod)

		prev := prod
		for f := range reservoirs[r].Formatter {
			form := make(chan []byte, 1)
			go reservoirs[r].Formatter[f].Format(prev, form)
			prev = form
		}

		go reservoirs[r].Consumer.Consume(prev)
	}

	done := make(chan struct{})
	<-done
}
