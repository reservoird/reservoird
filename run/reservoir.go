package run

import (
	"fmt"
	"sync"

	"github.com/reservoird/icd"
	"github.com/reservoird/reservoird/cfg"
)

// Reservoir is the structure for one reservoir flow
type Reservoir struct {
	Name         string
	ExpellerItem *ExpellerItem
	Disposed     bool
	Stopped      bool
	config       cfg.ReservoirCfg
	run          bool
	wg           *sync.WaitGroup
}

// NewReservoir setups the flow for one reservoir flow
func NewReservoir(config cfg.ReservoirCfg) (*Reservoir, error) {
	ings := make([]*IngesterItem, 0)
	for i := range config.ExpellerItem.IngesterItems {
		digs := make([]*DigesterItem, 0)
		for d := range config.ExpellerItem.IngesterItems[i].Digesters {
			digesterItem, err := NewDigesterItem(
				config.ExpellerItem.IngesterItems[i].Digesters[d].Location,
				config.ExpellerItem.IngesterItems[i].Digesters[d].Config,
				config.ExpellerItem.IngesterItems[i].Digesters[d].QueueItem.Location,
				config.ExpellerItem.IngesterItems[i].Digesters[d].QueueItem.Config,
			)
			if err != nil {
				return nil, err
			}
			digs = append(digs, digesterItem)
		}
		ingesterItem, err := NewIngesterItem(
			config.ExpellerItem.IngesterItems[i].Location,
			config.ExpellerItem.IngesterItems[i].Config,
			config.ExpellerItem.IngesterItems[i].QueueItem.Location,
			config.ExpellerItem.IngesterItems[i].QueueItem.Config,
			digs,
		)
		if err != nil {
			return nil, err
		}
		ings = append(ings, ingesterItem)
	}
	expellerItem, err := NewExpellerItem(
		config.ExpellerItem.Location,
		config.ExpellerItem.Config,
		ings,
	)
	if err != nil {
		return nil, err
	}
	reservoir := new(Reservoir)
	reservoir.Name = config.Name
	reservoir.ExpellerItem = expellerItem
	reservoir.config = config
	reservoir.Disposed = false
	reservoir.Stopped = true
	reservoir.run = false
	reservoir.wg = &sync.WaitGroup{}
	return reservoir, nil
}

// GetFlow returns the flow
func (o *Reservoir) GetFlow() ([]string, error) {
	if o.Disposed == true {
		return nil, fmt.Errorf("%s is disposed", o.Name)
	}
	flow := make([]string, 0)
	for i := range o.ExpellerItem.IngesterItems {
		flow = append(flow, o.ExpellerItem.IngesterItems[i].Ingester.Name())
		flow = append(flow, o.ExpellerItem.IngesterItems[i].QueueItem.Queue.Name())
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			flow = append(flow, o.ExpellerItem.IngesterItems[i].DigesterItems[d].Digester.Name())
			flow = append(flow, o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue.Name())
		}
	}
	flow = append(flow, o.ExpellerItem.Expeller.Name())
	return flow, nil
}

// Start starts system
func (o *Reservoir) Start() error {
	if o.Disposed == true {
		return fmt.Errorf("%s is disposed", o.Name)
	}
	if o.Stopped == false {
		return fmt.Errorf("%s is already started", o.Name)
	}
	o.Stopped = false
	var prevQueue icd.Queue
	expellerQueues := make([]icd.Queue, 0)
	for i := range o.ExpellerItem.IngesterItems {
		o.wg.Add(1)
		o.ExpellerItem.IngesterItems[i].QueueItem.mc.WaitGroup = o.wg
		go o.ExpellerItem.IngesterItems[i].QueueItem.Monitor()
		o.wg.Add(1)
		o.ExpellerItem.IngesterItems[i].mc.WaitGroup = o.wg
		go o.ExpellerItem.IngesterItems[i].Ingest()
		prevQueue = o.ExpellerItem.IngesterItems[i].QueueItem.Queue
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			o.wg.Add(1)
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.mc.WaitGroup = o.wg
			go o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Monitor()
			o.wg.Add(1)
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].mc.WaitGroup = o.wg
			go o.ExpellerItem.IngesterItems[i].DigesterItems[d].Digest(prevQueue)
			prevQueue = o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue
		}
		expellerQueues = append(expellerQueues, prevQueue)
	}
	o.wg.Add(1)
	o.ExpellerItem.mc.WaitGroup = o.wg
	go o.ExpellerItem.Expel(expellerQueues)
	return nil
}

// initStop initiates stop sequence
func (o *Reservoir) initStop() {
	o.ExpellerItem.mc.DoneChan <- struct{}{}
	for i := range o.ExpellerItem.IngesterItems {
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.mc.DoneChan <- struct{}{}
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].mc.DoneChan <- struct{}{}
		}
		o.ExpellerItem.IngesterItems[i].QueueItem.mc.DoneChan <- struct{}{}
		o.ExpellerItem.IngesterItems[i].mc.DoneChan <- struct{}{}
	}
}

// wait waits for system to stop
func (o *Reservoir) wait() {
	o.wg.Wait()
	o.Stopped = true
}

// Stop stops and waits
func (o *Reservoir) Stop() error {
	if o.Disposed == true {
		return fmt.Errorf("%s is disposed", o.Name)
	}
	if o.Stopped == true {
		return fmt.Errorf("%s is already stopped", o.Name)
	}
	o.initStop()
	o.wait()
	return nil
}

// Dispose disposes reservoir
func (o *Reservoir) Dispose() error {
	if o.Stopped == false {
		return fmt.Errorf("%s is running", o.Name)
	}
	o.Disposed = true
	return nil
}
