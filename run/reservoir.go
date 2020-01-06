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
	return reservoir, nil
}

// GoFlow spawns the flow
func (o *Reservoir) GoFlow(wg *sync.WaitGroup) {
	var prevQueue icd.Queue
	expellerQueues := make([]icd.Queue, 0)
	for i := range o.ExpellerItem.IngesterItems {
		wg.Add(1)
		go o.ExpellerItem.IngesterItems[i].Ingest(
			o.ExpellerItem.IngesterItems[i].QueueItem.Queue,
			wg,
		)
		prevQueue = o.ExpellerItem.IngesterItems[i].QueueItem.Queue
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			wg.Add(1)
			go o.ExpellerItem.IngesterItems[i].DigesterItems[d].Digest(
				prevQueue,
				o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue,
				wg,
			)
			prevQueue = o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue
		}
		expellerQueues = append(expellerQueues, prevQueue)
	}
	wg.Add(1)
	go o.ExpellerItem.Expel(
		expellerQueues,
		wg,
	)
}

// GoMonitor spaws the monitor
func (o *Reservoir) GoMonitor(wg *sync.WaitGroup) {
	for i := range o.ExpellerItem.IngesterItems {
		wg.Add(1)
		go o.ExpellerItem.IngesterItems[i].QueueItem.Monitor(wg)
		wg.Add(1)
		go o.ExpellerItem.IngesterItems[i].Monitor(wg)
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			wg.Add(1)
			go o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Monitor(wg)
			wg.Add(1)
			go o.ExpellerItem.IngesterItems[i].DigesterItems[d].Monitor(wg)
		}
	}
	wg.Add(1)
	go o.ExpellerItem.Monitor(wg)
}
