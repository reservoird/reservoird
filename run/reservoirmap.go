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
	Map  map[string]*Reservoir
	lock *sync.Mutex
}

// NewReservoirMap setups the flow
func NewReservoirMap(rsv cfg.Cfg) (*ReservoirMap, error) {
	o := new(ReservoirMap)
	o.Map = make(map[string]*Reservoir)
	o.lock = &sync.Mutex{}
	for r := range rsv.Reservoirs {
		reservoir, err := NewReservoir(rsv.Reservoirs[r])
		if err != nil {
			return nil, err
		}
		o.Map[reservoir.Name] = reservoir
	}
	return o, nil
}

// StartAll start system
func (o *ReservoirMap) StartAll() {
	o.lock.Lock()
	defer o.lock.Unlock()

	for _, reservoir := range o.Map {
		reservoir.Start()
	}
}

// UpdateAll updates stats
func (o *ReservoirMap) UpdateAll() {
	o.lock.Lock()
	defer o.lock.Unlock()

	for _, reservoir := range o.Map {
		reservoir.Update()
	}
}

// InitStopAll stops system
func (o *ReservoirMap) InitStopAll() {
	o.lock.Lock()
	defer o.lock.Unlock()

	for _, reservoir := range o.Map {
		reservoir.InitStop()
	}
}

// WaitAll waits
func (o *ReservoirMap) WaitAll() {
	o.lock.Lock()
	defer o.lock.Unlock()

	for _, reservoir := range o.Map {
		reservoir.Wait()
	}
}

// StopAll stops system
func (o *ReservoirMap) StopAll() {
	o.lock.Lock()
	defer o.lock.Unlock()

	for _, reservoir := range o.Map {
		reservoir.InitStop()
	}

	for _, reservoir := range o.Map {
		reservoir.Wait()
	}
}

// Start start system
func (o *ReservoirMap) Start(name string) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	reservoir, ok := o.Map[name]
	if ok == false {
		return fmt.Errorf("%s: no reservoir found", name)
	}
	err := reservoir.Start()
	if err != nil {
		return err
	}
	return nil
}

// InitStop stop system
func (o *ReservoirMap) InitStop(name string) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	reservoir, ok := o.Map[name]
	if ok == false {
		return fmt.Errorf("%s: no reservoir found", name)
	}
	err := reservoir.InitStop()
	if err != nil {
		return err
	}
	return nil
}

// Wait waits
func (o *ReservoirMap) Wait(name string) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	reservoir, ok := o.Map[name]
	if ok == false {
		return fmt.Errorf("%s: no reservoir found", name)
	}
	reservoir.Wait()
	return nil
}

// Stop stop system
func (o *ReservoirMap) Stop(name string) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	reservoir, ok := o.Map[name]
	if ok == false {
		return fmt.Errorf("%s: no reservoir found", name)
	}
	err := reservoir.InitStop()
	if err != nil {
		return err
	}
	reservoir.Wait()
	return nil
}

// Retrieve stop system
func (o *ReservoirMap) Retrieve(name string) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	reservoir, ok := o.Map[name]
	if ok == false {
		return fmt.Errorf("%s: no reservoir found", name)
	}
	if reservoir.Disposed() == false {
		return fmt.Errorf("%s: already created", name)
	}
	err := reservoir.Retrieve()
	if err != nil {
		return err
	}
	return nil
}

// Dispose stop system
func (o *ReservoirMap) Dispose(name string) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	reservoir, ok := o.Map[name]
	if ok == false {
		return fmt.Errorf("%s: no reservoir found", name)
	}
	if reservoir.Disposed() == true {
		return fmt.Errorf("%s: already disposed", name)
	}
	err := reservoir.Dispose()
	if err != nil {
		return err
	}
	reservoir.Wait()
	return nil
}

// GetReservoirs gets reservoirs
func (o *ReservoirMap) GetReservoirs() map[string][]interface{} {
	o.lock.Lock()
	defer o.lock.Unlock()

	reservoirMap := make(map[string][]interface{})
	for _, reservoir := range o.Map {
		r, err := reservoir.GetReservoir()
		if err == nil {
			reservoirMap[reservoir.Name] = r

		}
	}
	return reservoirMap
}

// GetReservoir gets reservoir
func (o *ReservoirMap) GetReservoir(name string) []interface{} {
	o.lock.Lock()
	defer o.lock.Unlock()

	reservoir, ok := o.Map[name]
	if ok == false {
		return nil
	}
	r, err := reservoir.GetReservoir()
	if err != nil {
		return nil
	}
	return r
}

// GetFlows gets flows
func (o *ReservoirMap) GetFlows() map[string][]string {
	o.lock.Lock()
	defer o.lock.Unlock()

	flows := make(map[string][]string)
	for _, reservoir := range o.Map {
		flow, err := reservoir.GetFlow()
		if err == nil {
			flows[reservoir.Name] = flow
		}
	}
	return flows
}

// GetFlow gets flows
func (o *ReservoirMap) GetFlow(name string) []string {
	o.lock.Lock()
	defer o.lock.Unlock()

	reservoir, ok := o.Map[name]
	if ok == false {
		return nil
	}
	flow, err := reservoir.GetFlow()
	if err != nil {
		return nil
	}
	return flow
}
