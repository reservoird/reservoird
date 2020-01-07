package run

import (
	"sync"

	"github.com/reservoird/icd"
	"github.com/reservoird/reservoird/cfg"
)

// Reservoir is the structure for one reservoir flow
type Reservoir struct {
	Name         string
	ExpellerItem *ExpellerItem
	wgFlow       *sync.WaitGroup
	wgMonitor    *sync.WaitGroup
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
	reservoir.wgFlow = &sync.WaitGroup{}
	reservoir.wgMonitor = &sync.WaitGroup{}
	return reservoir, nil
}

// GoFlow spawns the flow
func (o *Reservoir) GoFlow() {
	var prevQueue icd.Queue
	expellerQueues := make([]icd.Queue, 0)
	for i := range o.ExpellerItem.IngesterItems {
		o.wgFlow.Add(1)
		go o.ExpellerItem.IngesterItems[i].Ingest(o.wgFlow)
		prevQueue = o.ExpellerItem.IngesterItems[i].QueueItem.Queue
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			o.wgFlow.Add(1)
			go o.ExpellerItem.IngesterItems[i].DigesterItems[d].Digest(prevQueue, o.wgFlow)
			prevQueue = o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue
		}
		expellerQueues = append(expellerQueues, prevQueue)
	}
	o.wgFlow.Add(1)
	go o.ExpellerItem.Expel(expellerQueues, o.wgFlow)
}

// GoMonitor spaws the monitor
func (o *Reservoir) GoMonitor() {
	for i := range o.ExpellerItem.IngesterItems {
		o.wgMonitor.Add(1)
		go o.ExpellerItem.IngesterItems[i].QueueItem.Monitor(o.wgMonitor)
		o.wgMonitor.Add(1)
		go o.ExpellerItem.IngesterItems[i].Monitor(o.wgMonitor)
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			o.wgMonitor.Add(1)
			go o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Monitor(o.wgMonitor)
			o.wgMonitor.Add(1)
			go o.ExpellerItem.IngesterItems[i].DigesterItems[d].Monitor(o.wgMonitor)
		}
	}
	o.wgMonitor.Add(1)
	go o.ExpellerItem.Monitor(o.wgMonitor)
}

// StopFlow initites the end of the flow
func (o *Reservoir) StopFlow() {
	o.ExpellerItem.flowDoneChan <- struct{}{}
	for i := range o.ExpellerItem.IngesterItems {
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].flowDoneChan <- struct{}{}
		}
		o.ExpellerItem.IngesterItems[i].flowDoneChan <- struct{}{}
	}
}

// WaitFlow waits for flow to stop
func (o *Reservoir) WaitFlow() {
	o.wgFlow.Wait()
}

// StopMonitor initites the end of the monitor
func (o *Reservoir) StopMonitor() {
	o.ExpellerItem.monitorDoneChan <- struct{}{}
	for i := range o.ExpellerItem.IngesterItems {
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.monitorDoneChan <- struct{}{}
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].monitorDoneChan <- struct{}{}
		}
		o.ExpellerItem.IngesterItems[i].QueueItem.monitorDoneChan <- struct{}{}
		o.ExpellerItem.IngesterItems[i].monitorDoneChan <- struct{}{}
	}
}

// WaitMonitor waits for monitor to stop
func (o *Reservoir) WaitMonitor() {
	o.wgMonitor.Wait()
}
