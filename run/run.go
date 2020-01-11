package run

import (
	"github.com/reservoird/reservoird/cfg"
	log "github.com/sirupsen/logrus"
)

// Constants used for map index
const (
	Queues    = 0
	Ingesters = 1
	Digesters = 2
	Expellers = 3
)

// Reservoirs contains all reservoirs
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

// Start start system
func (o *Reservoirs) Start() {
	for r := range o.Reservoirs {
		o.Reservoirs[r].Start()
	}
}

// Stop stops system
func (o *Reservoirs) Stop() {
	log.Debug("into reservoirs.stop")
	for r := range o.Reservoirs {
		o.Reservoirs[r].Stop()
	}
	log.Debug("outof reservoirs.stop")
}

// Wait waits system to stop
func (o *Reservoirs) Wait() {
	log.Debug("into reservoirs.wait")
	for r := range o.Reservoirs {
		o.Reservoirs[r].wait()
	}
	log.Debug("outof reservoirs.wait")
}

// Cleanup stops and waits for flows and monitors
func (o *Reservoirs) Cleanup() {
	log.Debug("into reservoirs.cleanup")
	o.Stop()
	o.Wait()
	log.Debug("outof reservoirs.cleanup")
}

// Run runs the setup
func (o *Reservoirs) Run() error {
	o.Start()
	return nil
}
