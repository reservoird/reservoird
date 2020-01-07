package run

import (
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
func (o *Reservoirs) GoFlows() {
	for r := range o.Reservoirs {
		o.Reservoirs[r].GoFlow()
	}
}

// GoMonitors spawns all monitors
func (o *Reservoirs) GoMonitors() {
	for r := range o.Reservoirs {
		o.Reservoirs[r].GoMonitor()
	}
}

// StopFlows stops all flows
func (o *Reservoirs) StopFlows() {
	for r := range o.Reservoirs {
		o.Reservoirs[r].StopFlow()
	}
}

// WaitFlows waits for all flows to stop
func (o *Reservoirs) WaitFlows() {
	for r := range o.Reservoirs {
		o.Reservoirs[r].WaitFlow()
	}
}

// StopMonitors stops all monitors
func (o *Reservoirs) StopMonitors() {
	for r := range o.Reservoirs {
		o.Reservoirs[r].StopMonitor()
	}
}

// WaitMonitors waits for all monitors to stop
func (o *Reservoirs) WaitMonitors() {
	for r := range o.Reservoirs {
		o.Reservoirs[r].WaitMonitor()
	}
}

// Cleanup stops and waits for flows and monitors
func (o *Reservoirs) Cleanup() {
	o.StopFlows()
	o.WaitFlows()
	o.StopMonitors()
	o.WaitMonitors()
}

// Run runs the setup
func (o *Reservoirs) Run() error {
	o.GoFlows()
	o.GoMonitors()
	return nil
}
