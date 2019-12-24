package run

import (
	"fmt"
	"plugin"

	"github.com/reservoird/reservoird/cfg"
)

// IngesterItem is what is needed to run an ingester
type IngesterItem struct {
	ConfigFile  string
	ChannelSize int
	Ingester    Ingester
}

// DigesterItem is what is needed to run a digester
type DigesterItem struct {
	ConfigFile  string
	ChannelSize int
	Digester    Digester
}

// ExpellerItem is what is needed to run an expeller
type ExpellerItem struct {
	ConfigFile string
	Expeller   Expeller
}

// Reservoir is the structure of the reservoir flow
type Reservoir struct {
	IngesterItem  IngesterItem
	DigesterItems []DigesterItem
	ExpellerItem  ExpellerItem
}

// NewReservoirs setups the flow
func NewReservoirs(rsv cfg.Cfg) ([]Reservoir, error) {
	reservoirs := make([]Reservoir, 0)
	for r := range rsv.Reservoirs {
		// setup ingester
		ingester, err := plugin.Open(rsv.Reservoirs[r].Ingester.Location)
		if err != nil {
			return nil, err
		}
		symbolIngester, err := ingester.Lookup("Ingester")
		if err != nil {
			return nil, err
		}
		ingest, ok := symbolIngester.(Ingester)
		if ok == false {
			return nil, fmt.Errorf("error Ingester interface not implemented")
		}
		ing := IngesterItem{
			ConfigFile:  rsv.Reservoirs[r].Ingester.ConfigFile,
			ChannelSize: rsv.Reservoirs[r].Ingester.ChannelSize,
			Ingester:    ingest,
		}

		// setup digesters
		digs := make([]DigesterItem, 0)
		for d := range rsv.Reservoirs[r].Digesters {
			digester, err := plugin.Open(rsv.Reservoirs[r].Digesters[d].Location)
			if err != nil {
				return nil, err
			}
			symbolDigester, err := digester.Lookup("Digester")
			if err != nil {
				return nil, err
			}
			digest, ok := symbolDigester.(Digester)
			if ok == false {
				return nil, fmt.Errorf("error Digester interface not implemented")
			}
			dig := DigesterItem{
				ConfigFile:  rsv.Reservoirs[r].Digesters[d].ConfigFile,
				ChannelSize: rsv.Reservoirs[r].Digesters[d].ChannelSize,
				Digester:    digest,
			}
			digs = append(digs, dig)
		}

		// setup expellers
		expeller, err := plugin.Open(rsv.Reservoirs[r].Expeller.Location)
		if err != nil {
			return nil, err
		}
		symbolExpeller, err := expeller.Lookup("Expeller")
		if err != nil {
			return nil, err
		}
		expel, ok := symbolExpeller.(Expeller)
		if ok == false {
			return nil, fmt.Errorf("error Expeller interface not implemented")
		}
		exp := ExpellerItem{
			ConfigFile: rsv.Reservoirs[r].Expeller.ConfigFile,
			Expeller:   expel,
		}

		reservoirs = append(reservoirs,
			Reservoir{
				IngesterItem:  ing,
				DigesterItems: digs,
				ExpellerItem:  exp,
			},
		)
	}
	return reservoirs, nil
}

// Cfg setups configuration
func Cfg(reservoirs []Reservoir) error {
	for r := range reservoirs {
		ingcfg := reservoirs[r].IngesterItem.ConfigFile
		err := reservoirs[r].IngesterItem.Ingester.Config(ingcfg)
		if err != nil {
			return err
		}
		for d := range reservoirs[r].DigesterItems {
			digcfg := reservoirs[r].DigesterItems[d].ConfigFile
			err := reservoirs[r].DigesterItems[d].Digester.Config(digcfg)
			if err != nil {
				return err
			}
		}
		expcfg := reservoirs[r].ExpellerItem.ConfigFile
		err = reservoirs[r].ExpellerItem.Expeller.Config(expcfg)
		if err != nil {
			return err
		}
	}
	return nil
}

// Run runs the setup
func Run(reservoirs []Reservoir) {
	for r := range reservoirs {
		ingchan := make(chan []byte, reservoirs[r].IngesterItem.ChannelSize)
		go reservoirs[r].IngesterItem.Ingester.Ingest(ingchan)

		prev := ingchan
		for d := range reservoirs[r].DigesterItems {
			digchan := make(chan []byte, reservoirs[r].DigesterItems[d].ChannelSize)
			go reservoirs[r].DigesterItems[d].Digester.Digest(prev, digchan)
			prev = digchan
		}

		go reservoirs[r].ExpellerItem.Expeller.Expel(prev)
	}

	// wait forever
	done := make(chan struct{})
	<-done
}
