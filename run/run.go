package run

import (
	"sync"

	"github.com/reservoird/reservoird/cfg"
)

// Constants used for map index
const (
	Queues    = 0
	Ingesters = 1
	Digesters = 2
	Expellers = 3
)

type Reservoirs struct {
	Reservoirs map[string]*Reservoir
}

// NewReservoirs setups the flow
func NewReservoirs(rsv cfg.Cfg) (*Reservoirs, error) {
	o := new(Reservoirs)
	o.Reservoirs = make(map[string]*Reservoir)
	for r := range rsv.Reservoirs {
		reservoir, err := NewReservoir(rsv.Reservoirs[r])
		if err != nil {
			return nil, err
		}
		o.Reservoirs[reservoir.Name] = reservoir
	}
	return o, nil
}

// GoFlows spawns all flows
func (o *Reservoirs) GoFlows(wg *sync.WaitGroup) {
	for r := range o.Reservoirs {
		o.Reservoirs[r].GoFlow(wg)
	}
}

// GoMonitors spawns all monitors
func (o *Reservoirs) GoMonitors(wg *sync.WaitGroup) {
	for r := range o.Reservoirs {
		o.Reservoirs[r].GoMonitor(wg)
	}
}

// Run runs the setup
func (o *Reservoirs) Run() error {
	wg := &sync.WaitGroup{}
	o.GoFlows(wg)
	o.GoMonitors(wg)

	server, err := NewServer(o)
	if err != nil {
		return err
	}

	wg.Add(1)
	go server.Monitor(wg)

	err = server.Serve()
	if err != nil {
		return err
	}

	wg.Wait()
	return nil
}
