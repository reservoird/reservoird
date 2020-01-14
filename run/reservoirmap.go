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
	Map      map[string]*Reservoir
	Disposed map[string]bool
	Stopped  map[string]bool
	lock     *sync.Mutex
}

// NewReservoirMap setups the flow
func NewReservoirMap(rsv cfg.Cfg) (*ReservoirMap, error) {
	o := new(ReservoirMap)
	o.Map = make(map[string]*Reservoir)
	o.Disposed = make(map[string]bool)
	o.Stopped = make(map[string]bool)
	o.lock = &sync.Mutex{}
	for r := range rsv.Reservoirs {
		reservoir, err := NewReservoir(rsv.Reservoirs[r])
		if err != nil {
			return nil, err
		}
		o.Map[reservoir.Name] = reservoir
		o.Disposed[reservoir.Name] = false
		o.Stopped[reservoir.Name] = true
	}
	return o, nil
}

// StartAll start system
func (o *ReservoirMap) StartAll() {
	o.lock.Lock()
	defer o.lock.Unlock()

	for name := range o.Map {
		if o.Disposed[name] == false && o.Stopped[name] == true {
			o.Map[name].Start()
			o.Stopped[name] = false
		}
	}
}

// UpdateAll updates stats
func (o *ReservoirMap) UpdateAll() {
	o.lock.Lock()
	defer o.lock.Unlock()

	for name := range o.Map {
		if o.Stopped[name] == false {
			o.Map[name].Update()
		}
	}
}

// UpdateFinalAll updates stats
func (o *ReservoirMap) UpdateFinalAll() {
	o.lock.Lock()
	defer o.lock.Unlock()

	for name := range o.Map {
		o.Map[name].UpdateFinal()
	}
}

// InitStopAll stops system
func (o *ReservoirMap) InitStopAll() {
	o.lock.Lock()
	defer o.lock.Unlock()

	for name := range o.Map {
		if o.Disposed[name] == false && o.Stopped[name] == false {
			o.Map[name].InitStop()
			o.Stopped[name] = true
		}
	}
}

// WaitAll waits
func (o *ReservoirMap) WaitAll() {
	o.lock.Lock()
	defer o.lock.Unlock()

	for name := range o.Map {
		o.Map[name].Wait()
	}
}

// StopAll stops system
func (o *ReservoirMap) StopAll() {
	o.lock.Lock()
	defer o.lock.Unlock()

	for name := range o.Map {
		o.Map[name].InitStop()
	}

	o.UpdateAll()

	for name := range o.Map {
		o.Map[name].Wait()
	}
}

// StoppedAll stops system
func (o *ReservoirMap) StoppedAll() bool {
	o.lock.Lock()
	defer o.lock.Unlock()

	stopped := true
	for name := range o.Map {
		if o.Stopped[name] == false {
			stopped = false
			break
		}
	}
	return stopped
}

// Start start system
func (o *ReservoirMap) Start(name string) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	reservoir, ok := o.Map[name]
	if ok == false {
		return fmt.Errorf("%s: no reservoir found", name)
	}
	if o.Disposed[name] == true {
		return fmt.Errorf("%s: disposed", name)
	}
	if o.Stopped[name] == false {
		return fmt.Errorf("%s: already running", name)
	}
	err := reservoir.Start()
	if err != nil {
		return err
	}
	o.Stopped[name] = false
	return nil
}

// UpdateFinal updates stats
func (o *ReservoirMap) UpdateFinal(name string) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	reservoir, ok := o.Map[name]
	if ok == false {
		return fmt.Errorf("%s: no reservoir found", name)
	}
	reservoir.UpdateFinal()
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
	if o.Disposed[name] == true {
		return fmt.Errorf("%s: disposed", name)
	}
	if o.Stopped[name] == true {
		return fmt.Errorf("%s: already stopped", name)
	}
	err := reservoir.InitStop()
	if err != nil {
		return err
	}
	o.Stopped[name] = true
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

// UpdateFinalAndWait waits
func (o *ReservoirMap) UpdateFinalAndWait(name string) error {
	err := o.UpdateFinal(name)
	if err != nil {
		return err
	}
	err = o.Wait(name)
	if err != nil {
		return err
	}
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

	_, ok := o.Map[name]
	if ok == false {
		return fmt.Errorf("%s: no reservoir found", name)
	}
	o.Disposed[name] = false
	return nil
}

// Dispose stop system
func (o *ReservoirMap) Dispose(name string) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	_, ok := o.Map[name]
	if ok == false {
		return fmt.Errorf("%s: no reservoir found", name)
	}
	if o.Disposed[name] == true {
		return fmt.Errorf("%s: already disposed", name)
	}
	if o.Stopped[name] == false {
		return fmt.Errorf("%s: running", name)
	}
	o.Disposed[name] = true
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
			if o.Disposed[reservoir.Name] == false {
				reservoirMap[reservoir.Name] = r
			}
		}
	}
	return reservoirMap
}

// GetReservoir gets reservoir
func (o *ReservoirMap) GetReservoir(name string) ([]interface{}, bool, bool) {
	o.lock.Lock()
	defer o.lock.Unlock()

	reservoir, ok := o.Map[name]
	if ok == false {
		return nil, false, false
	}
	if o.Disposed[name] == true {
		return nil, false, false
	}
	r, err := reservoir.GetReservoir()
	if err != nil {
		return nil, false, false
	}
	return r, o.Stopped[name], o.Disposed[name]
}

// GetFlows gets flows
func (o *ReservoirMap) GetFlows() map[string][]string {
	o.lock.Lock()
	defer o.lock.Unlock()

	flows := make(map[string][]string)
	for _, reservoir := range o.Map {
		flow, err := reservoir.GetFlow()
		if err == nil {
			if o.Disposed[reservoir.Name] == false {
				flows[reservoir.Name] = flow
			}
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
	if o.Disposed[name] == true {
		return nil
	}
	flow, err := reservoir.GetFlow()
	if err != nil {
		return nil
	}
	return flow
}
