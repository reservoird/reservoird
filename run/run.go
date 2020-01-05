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

// NewReservoirs setups the flow
func NewReservoirs(rsv cfg.Cfg) (map[string]*Reservoir, error) {
	reservoirs := make(map[string]*Reservoir)
	for r := range rsv.Reservoirs {
		reservoir, err := NewReservoir(rsv.Reservoirs[r])
		if err != nil {
			return nil, err
		}
		reservoirs[reservoir.Name] = reservoir
	}
	return reservoirs, nil
}

// GoFlows spawns all flows
func GoFlows(reservoirs map[string]*Reservoir, wg *sync.WaitGroup) {
	for r := range reservoirs {
		reservoirs[r].GoFlow(wg)
	}
}

// GoMonitors spawns all monitors
func GoMonitors(reservoirs map[string]*Reservoir, wg *sync.WaitGroup) {
	for r := range reservoirs {
		reservoirs[r].GoMonitor(wg)
	}
}

// Run runs the setup
func Run(reservoirs map[string]*Reservoir) error {
	wg := &sync.WaitGroup{}
	GoFlows(reservoirs, wg)
	GoMonitors(reservoirs, wg)

	server, err := NewServer(reservoirs)
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
