package run

import (
	"fmt"
	"plugin"

	"github.com/reservoird/reservoird/cfg"
)

// ProducerItem is what is needed to run a producer
type ProducerItem struct {
	ConfigFile  string
	ChannelSize int
	Producer    Producer
}

// FormatterItem is what is needed to run a formatter
type FormatterItem struct {
	ConfigFile  string
	ChannelSize int
	Formatter   Formatter
}

// ConsumerItem is what is needed to run a consumer
type ConsumerItem struct {
	ConfigFile string
	Consumer   Consumer
}

// Reservoir is the structure of the reservoir flow
type Reservoir struct {
	ProducerItem   ProducerItem
	FormatterItems []FormatterItem
	ConsumerItem   ConsumerItem
}

// NewReservoirs setups the flow
func NewReservoirs(rsv cfg.Cfg) ([]Reservoir, error) {
	reservoirs := make([]Reservoir, 0)
	for r := range rsv.Reservoirs {
		// setup producer
		producer, err := plugin.Open(rsv.Reservoirs[r].Producer.Location)
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
		pi := ProducerItem{
			ConfigFile:  rsv.Reservoirs[r].Producer.ConfigFile,
			ChannelSize: rsv.Reservoirs[r].Producer.ChannelSize,
			Producer:    prod,
		}

		// setup formatters
		fis := make([]FormatterItem, 0)
		for f := range rsv.Reservoirs[r].Formatters {
			formatter, err := plugin.Open(rsv.Reservoirs[r].Formatters[f].Location)
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
			fi := FormatterItem{
				ConfigFile:  rsv.Reservoirs[r].Formatters[f].ConfigFile,
				ChannelSize: rsv.Reservoirs[r].Formatters[f].ChannelSize,
				Formatter:   form,
			}
			fis = append(fis, fi)
		}

		// setup consumers
		consumer, err := plugin.Open(rsv.Reservoirs[r].Consumer.Location)
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
		ci := ConsumerItem{
			ConfigFile: rsv.Reservoirs[r].Consumer.ConfigFile,
			Consumer:   cons,
		}

		reservoirs = append(reservoirs,
			Reservoir{
				ProducerItem:   pi,
				FormatterItems: fis,
				ConsumerItem:   ci,
			},
		)
	}
	return reservoirs, nil
}

// Cfg setups configuration
func Cfg(reservoirs []Reservoir) error {
	for r := range reservoirs {
		pcfg := reservoirs[r].ProducerItem.ConfigFile
		err := reservoirs[r].ProducerItem.Producer.Config(pcfg)
		if err != nil {
			return err
		}
		for f := range reservoirs[r].FormatterItems {
			fcfg := reservoirs[r].FormatterItems[f].ConfigFile
			err := reservoirs[r].FormatterItems[f].Formatter.Config(fcfg)
			if err != nil {
				return err
			}
		}
		ccfg := reservoirs[r].ConsumerItem.ConfigFile
		err = reservoirs[r].ConsumerItem.Consumer.Config(ccfg)
		if err != nil {
			return err
		}
	}
	return nil
}

// Run runs the setup
func Run(reservoirs []Reservoir) {
	for r := range reservoirs {
		pchan := make(chan []byte, reservoirs[r].ProducerItem.ChannelSize)
		go reservoirs[r].ProducerItem.Producer.Produce(pchan)

		prev := pchan
		for f := range reservoirs[r].FormatterItems {
			fchan := make(chan []byte, reservoirs[r].FormatterItems[f].ChannelSize)
			go reservoirs[r].FormatterItems[f].Formatter.Format(prev, fchan)
			prev = fchan
		}

		go reservoirs[r].ConsumerItem.Consumer.Consume(prev)
	}

	// wait forever
	done := make(chan struct{})
	<-done
}
