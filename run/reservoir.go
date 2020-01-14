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
	config       cfg.ReservoirCfg
	run          bool
	disposed     bool
	stopped      bool
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
	reservoir.run = false
	reservoir.disposed = false
	reservoir.stopped = true
	reservoir.wg = &sync.WaitGroup{}
	return reservoir, nil
}

// GetReservoir return the reservoir
func (o *Reservoir) GetReservoir() ([]interface{}, error) {
	if o.disposed == true {
		return nil, fmt.Errorf("%s: disposed", o.Name)
	}
	reservoir := make([]interface{}, 0)
	for i := range o.ExpellerItem.IngesterItems {
		reservoir = append(reservoir, o.ExpellerItem.IngesterItems[i].stats)
		reservoir = append(reservoir, o.ExpellerItem.IngesterItems[i].QueueItem.stats)
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			reservoir = append(reservoir, o.ExpellerItem.IngesterItems[i].DigesterItems[d].stats)
			reservoir = append(reservoir, o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.stats)
		}
		reservoir = append(reservoir, o.ExpellerItem.stats)
	}
	return reservoir, nil
}

// GetFlow returns the flow
func (o *Reservoir) GetFlow() ([]string, error) {
	if o.disposed == true {
		return nil, fmt.Errorf("%s: disposed", o.Name)
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
	if o.disposed == true {
		return fmt.Errorf("%s: disposed", o.Name)
	}
	if o.stopped == false {
		return fmt.Errorf("%s: already started", o.Name)
	}
	o.stopped = false
	var prevQueue icd.Queue
	expellerQueues := make([]icd.Queue, 0)
	for i := range o.ExpellerItem.IngesterItems {
		o.wg.Add(1)
		o.ExpellerItem.IngesterItems[i].QueueItem.MonitorControl.WaitGroup = o.wg
		go o.ExpellerItem.IngesterItems[i].QueueItem.Monitor()
		o.wg.Add(1)
		o.ExpellerItem.IngesterItems[i].MonitorControl.WaitGroup = o.wg
		go o.ExpellerItem.IngesterItems[i].Ingest()
		prevQueue = o.ExpellerItem.IngesterItems[i].QueueItem.Queue
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			o.wg.Add(1)
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.MonitorControl.WaitGroup = o.wg
			go o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Monitor()
			o.wg.Add(1)
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].MonitorControl.WaitGroup = o.wg
			go o.ExpellerItem.IngesterItems[i].DigesterItems[d].Digest(prevQueue)
			prevQueue = o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue
		}
		expellerQueues = append(expellerQueues, prevQueue)
	}
	o.wg.Add(1)
	o.ExpellerItem.MonitorControl.WaitGroup = o.wg
	go o.ExpellerItem.Expel(expellerQueues)
	return nil
}

// InitStop initiates a stop
func (o *Reservoir) InitStop() error {
	if o.disposed == true {
		return fmt.Errorf("%s: disposed", o.Name)
	}
	if o.stopped == true {
		return fmt.Errorf("%s: already stopped", o.Name)
	}
	o.ExpellerItem.MonitorControl.DoneChan <- struct{}{}
	for i := range o.ExpellerItem.IngesterItems {
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.MonitorControl.DoneChan <- struct{}{}
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].MonitorControl.DoneChan <- struct{}{}
		}
		o.ExpellerItem.IngesterItems[i].QueueItem.MonitorControl.DoneChan <- struct{}{}
		o.ExpellerItem.IngesterItems[i].MonitorControl.DoneChan <- struct{}{}
	}
	o.stopped = true
	return nil
}

// Stop stops and waits
func (o *Reservoir) Stop() error {
	err := o.InitStop()
	if err != nil {
		return err
	}
	o.wg.Wait()
	return nil
}

// Wait waits
func (o *Reservoir) Wait() {
	o.wg.Wait()
}

// Update updates stats
func (o *Reservoir) Update() error {
	if o.disposed == true {
		return fmt.Errorf("%s: disposed", o.Name)
	}

	for i := range o.ExpellerItem.IngesterItems {
		select {
		case stats := <-o.ExpellerItem.IngesterItems[i].MonitorControl.StatsChan:
			o.ExpellerItem.IngesterItems[i].stats = stats
		default:
		}
		select {
		case stats := <-o.ExpellerItem.IngesterItems[i].QueueItem.MonitorControl.StatsChan:
			o.ExpellerItem.IngesterItems[i].QueueItem.stats = stats
		default:
		}
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			select {
			case stats := <-o.ExpellerItem.IngesterItems[i].DigesterItems[d].MonitorControl.StatsChan:
				o.ExpellerItem.IngesterItems[i].DigesterItems[d].stats = stats
			default:
			}
			select {
			case stats := <-o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.MonitorControl.StatsChan:
				o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.stats = stats
			default:
			}
		}
	}
	select {
	case stats := <-o.ExpellerItem.MonitorControl.StatsChan:
		o.ExpellerItem.stats = stats
	default:
	}
	return nil
}

// Retrieve retains
func (o *Reservoir) Retrieve() error {
	o.disposed = false
	return nil
}

// Dispose disposes reservoir
func (o *Reservoir) Dispose() error {
	if o.stopped == false {
		return fmt.Errorf("%s: running", o.Name)
	}
	o.disposed = true
	return nil
}

// Stopped is stopped
func (o *Reservoir) Stopped() bool {
	return o.stopped
}

// Disposed is disposed
func (o *Reservoir) Disposed() bool {
	return o.disposed
}
