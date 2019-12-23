package run

import (
	"fmt"
	"plugin"

	"github.com/reservoird/reservoird/cfg"
)

// Reservoir is the structure of the reservoir flow
type Reservoir struct {
	ChannelSize int
	Producer    Producer
	Formatter   []Formatter
	Consumer    Consumer
}

// NewReservoirs setups the flow
func NewReservoirs(rsv cfg.Cfg) ([]Reservoir, error) {
	reservoirs := make([]Reservoir, 0)
	for r := range rsv.Reservoir {
		channelSize := rsv.Reservoir[r].ChannelSize
		// setup producer
		producer, err := plugin.Open(rsv.Reservoir[r].Producer)
		if err != nil {
			return nil, err
		}
		symbolProducer, err := producer.Lookup("Producer")
		if err != nil {
			return nil, err
		}
		prod, ok := symbolProducer.(Producer)
		if ok == false {
			return nil, fmt.Errorf("error Producer.Produce function not implemented")
		}

		// setup formatters
		forms := make([]Formatter, 0)
		for f := range rsv.Reservoir[r].Formatter {
			formatter, err := plugin.Open(rsv.Reservoir[r].Formatter[f])
			if err != nil {
				return nil, err
			}
			symbolFormatter, err := formatter.Lookup("Formatter")
			if err != nil {
				return nil, err
			}
			form, ok := symbolFormatter.(Formatter)
			if ok == false {
				return nil, fmt.Errorf("error Formatter.Format function not implemented")
			}
			forms = append(forms, form)
		}

		// setup consumers
		consumer, err := plugin.Open(rsv.Reservoir[r].Consumer)
		if err != nil {
			return nil, err
		}
		symbolConsumer, err := consumer.Lookup("Consumer")
		if err != nil {
			return nil, err
		}
		cons, ok := symbolConsumer.(Consumer)
		if ok == false {
			return nil, fmt.Errorf("error Consumer.Consume function not implemented")
		}

		reservoirs = append(reservoirs,
			Reservoir{
				ChannelSize: channelSize,
				Producer:    prod,
				Formatter:   forms,
				Consumer:    cons,
			},
		)
	}
	return reservoirs, nil
}

// Run runs the setup
func Run(reservoirs []Reservoir) {
	for r := range reservoirs {
		prod := make(chan []byte, reservoirs[r].ChannelSize)
		go reservoirs[r].Producer.Produce(prod)

		prev := prod
		for f := range reservoirs[r].Formatter {
			form := make(chan []byte, reservoirs[r].ChannelSize)
			go reservoirs[r].Formatter[f].Format(prev, form)
			prev = form
		}

		go reservoirs[r].Consumer.Consume(prev)
	}

	// wait forever
	done := make(chan struct{})
	<-done
}
