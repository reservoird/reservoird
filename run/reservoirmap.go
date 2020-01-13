package run

import (
	"fmt"
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

// ReservoirMap contains all reservoirs
type ReservoirMap struct {
	Map *sync.Map
}

// NewReservoirMap setups the flow
func NewReservoirMap(rsv cfg.Cfg) (*ReservoirMap, error) {
	o := new(ReservoirMap)
	o.Map = &sync.Map{}
	for r := range rsv.Reservoirs {
		reservoir, err := NewReservoir(rsv.Reservoirs[r])
		if err != nil {
			return nil, err
		}
		o.Map.Store(reservoir.Name, reservoir)
	}
	return o, nil
}

// StartAll start system
func (o *ReservoirMap) StartAll() {
	o.Map.Range(func(_, val interface{}) bool {
		r, ok := val.(*Reservoir)
		if ok == false {
			return true
		}
		r.Start()
		return true
	})
}

// UpdateAll updates stats
func (o *ReservoirMap) UpdateAll() {
	o.Map.Range(func(_, val interface{}) bool {
		r, ok := val.(*Reservoir)
		if ok == false {
			return true
		}
		r.Update()
		return true
	})
}

// InitStopAll stops system
func (o *ReservoirMap) InitStopAll() {
	o.Map.Range(func(_, val interface{}) bool {
		r, ok := val.(*Reservoir)
		if ok == false {
			return true
		}
		r.InitStop()
		return true
	})
}

// WaitAll waits
func (o *ReservoirMap) WaitAll() {
	o.Map.Range(func(_, val interface{}) bool {
		r, ok := val.(*Reservoir)
		if ok == false {
			return true
		}
		r.Wait()
		return true
	})
}

// StopAll stops system
func (o *ReservoirMap) StopAll() {
	o.InitStopAll()
	o.WaitAll()
}

// Start start system
func (o *ReservoirMap) Start(name string) error {
	val, ok := o.Map.Load(name)
	if ok == false {
		return fmt.Errorf("%s: no reservoir found", name)
	}
	r, ok := val.(*Reservoir)
	if ok == false {
		return fmt.Errorf("%s: internal error", name)
	}
	r.Start()
	return nil
}

// InitStop stop system
func (o *ReservoirMap) InitStop(name string) error {
	val, ok := o.Map.Load(name)
	if ok == false {
		return fmt.Errorf("%s: no reservoir found", name)
	}
	r, ok := val.(*Reservoir)
	if ok == false {
		return fmt.Errorf("%s: internal error", name)
	}
	r.InitStop()
	return nil
}

// Wait waits
func (o *ReservoirMap) Wait(name string) error {
	val, ok := o.Map.Load(name)
	if ok == false {
		return fmt.Errorf("%s: no reservoir found", name)
	}
	r, ok := val.(*Reservoir)
	if ok == false {
		return fmt.Errorf("%s: internal error", name)
	}
	r.Wait()
	return nil
}

// Stop stop system
func (o *ReservoirMap) Stop(name string) error {
	err := o.Stop(name)
	if err != nil {
		return err
	}
	err = o.Wait(name)
	if err != nil {
		return err
	}
	return nil
}

// GetReservoirs gets reservoirs
func (o *ReservoirMap) GetReservoirs() map[string]*Reservoir {
	reservoirs := make(map[string]*Reservoir)
	o.Map.Range(func(_, val interface{}) bool {
		r, ok := val.(*Reservoir)
		if ok == false {
			return true
		}
		reservoirs[r.Name] = r
		return true
	})
	return reservoirs
}

// GetReservoir gets reservoir
func (o *ReservoirMap) GetReservoir(name string) *Reservoir {
	r, ok := o.Map.Load(name)
	if ok == false {
		return nil
	}
	reservoir, ok := r.(*Reservoir)
	if ok == false {
		return nil
	}
	return reservoir
}

// GetFlows gets flows
func (o *ReservoirMap) GetFlows() map[string][]string {
	flows := make(map[string][]string)
	o.Map.Range(func(_, val interface{}) bool {
		r, ok := val.(*Reservoir)
		if ok == false {
			return true
		}
		flow, err := r.GetFlow()
		if err != nil {
			return true
		}
		flows[r.Name] = flow
		return true
	})
	return flows
}

// GetFlow gets flows
func (o *ReservoirMap) GetFlow(name string) map[string][]string {
	val, ok := o.Map.Load(name)
	if ok == false {
		return nil
	}
	r, ok := val.(*Reservoir)
	if ok == false {
		return nil
	}
	flow, err := r.GetFlow()
	if err != nil {
		return nil
	}
	flows := map[string][]string{
		name: flow,
	}
	return flows
}
