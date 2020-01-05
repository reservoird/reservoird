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
		go o.ExpellerItem.IngesterItems[i].Ingester.Ingest(
			o.ExpellerItem.IngesterItems[i].QueueItem.Queue,
			o.ExpellerItem.IngesterItems[i].flowDoneChan,
			wg,
		)
		prevQueue = o.ExpellerItem.IngesterItems[i].QueueItem.Queue
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			wg.Add(1)
			go o.ExpellerItem.IngesterItems[i].DigesterItems[d].Digester.Digest(
				prevQueue,
				o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue,
				o.ExpellerItem.IngesterItems[i].DigesterItems[d].flowDoneChan,
				wg,
			)
			prevQueue = o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue
		}
		expellerQueues = append(expellerQueues, prevQueue)
	}
	wg.Add(1)
	go o.ExpellerItem.Expeller.Expel(
		expellerQueues,
		o.ExpellerItem.flowDoneChan,
		wg,
	)
}

// GoMonitor spaws the monitor
func (o *Reservoir) GoMonitor(wg *sync.WaitGroup) {
	for i := range o.ExpellerItem.IngesterItems {
		wg.Add(1)
		go o.ExpellerItem.IngesterItems[i].QueueItem.Queue.Monitor(
			o.ExpellerItem.IngesterItems[i].QueueItem.monitorStatsChan,
			o.ExpellerItem.IngesterItems[i].QueueItem.monitorClearChan,
			o.ExpellerItem.IngesterItems[i].QueueItem.monitorDoneChan,
			wg,
		)
		wg.Add(1)
		go o.ExpellerItem.IngesterItems[i].Ingester.Monitor(
			o.ExpellerItem.IngesterItems[i].monitorStatsChan,
			o.ExpellerItem.IngesterItems[i].monitorClearChan,
			o.ExpellerItem.IngesterItems[i].monitorDoneChan,
			wg,
		)
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			wg.Add(1)
			go o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue.Monitor(
				o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.monitorStatsChan,
				o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.monitorClearChan,
				o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.monitorDoneChan,
				wg,
			)
			wg.Add(1)
			go o.ExpellerItem.IngesterItems[i].DigesterItems[d].Digester.Monitor(
				o.ExpellerItem.IngesterItems[i].DigesterItems[d].monitorStatsChan,
				o.ExpellerItem.IngesterItems[i].DigesterItems[d].monitorClearChan,
				o.ExpellerItem.IngesterItems[i].DigesterItems[d].monitorDoneChan,
				wg,
			)
		}
	}
	wg.Add(1)
	go o.ExpellerItem.Expeller.Monitor(
		o.ExpellerItem.monitorStatsChan,
		o.ExpellerItem.monitorClearChan,
		o.ExpellerItem.monitorDoneChan,
		wg,
	)
}
