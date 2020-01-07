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
	wg         *sync.WaitGroup
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
	o.wg = &sync.WaitGroup{}
	return o, nil
}

// GoFlows spawns all flows
func (o *Reservoirs) GoFlows() {
	for r := range o.Reservoirs {
		o.Reservoirs[r].GoFlow(o.wg)
	}
}

// GoMonitors spawns all monitors
func (o *Reservoirs) GoMonitors() {
	for r := range o.Reservoirs {
		o.Reservoirs[r].GoMonitor(o.wg)
	}
}

// StopFlows stops all flows
func (o *Reservoirs) StopFlows() {
	for r := range o.Reservoirs {
		o.Reservoirs[r].StopFlow()
	}
}

// StopMonitors stops all monitors
func (o *Reservoirs) StopMonitors() {
	for r := range o.Reservoirs {
		o.Reservoirs[r].StopMonitor()
	}
}

// Run runs the setup
func (o *Reservoirs) Run() error {
	o.GoFlows()
	o.GoMonitors()

	server, err := NewServer(o)
	if err != nil {
		return err
	}

	o.wg.Add(1)
	go server.Monitor(o.wg)

	err = server.Serve()
	if err != nil {
		return err
	}

	o.wg.Wait()
	return nil
}
