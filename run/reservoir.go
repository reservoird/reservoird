package run

import (
	"sync"

	"github.com/reservoird/icd"
	"github.com/reservoird/proxy"
	"github.com/reservoird/reservoird/cfg"
)

// Reservoir is the structure for one reservoir flow
type Reservoir struct {
	Name         string
	ExpellerItem *ExpellerItem
	config       cfg.ReservoirCfg
	run          bool
	wg           *sync.WaitGroup
}

// NewReservoir setups the flow for one reservoir flow
func NewReservoir(
	config cfg.ReservoirCfg,
	plugin proxy.Plugin,
) (*Reservoir, error) {
	ings := make([]*IngesterItem, 0)
	for i := range config.ExpellerItem.IngesterItems {
		digs := make([]*DigesterItem, 0)
		for d := range config.ExpellerItem.IngesterItems[i].Digesters {
			digesterItem, err := NewDigesterItem(
				config.ExpellerItem.IngesterItems[i].Digesters[d].Location,
				config.ExpellerItem.IngesterItems[i].Digesters[d].Config,
				config.ExpellerItem.IngesterItems[i].Digesters[d].QueueItem.Location,
				config.ExpellerItem.IngesterItems[i].Digesters[d].QueueItem.Config,
				plugin,
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
			plugin,
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
		plugin,
	)
	if err != nil {
		return nil, err
	}
	reservoir := new(Reservoir)
	reservoir.Name = config.Name
	reservoir.ExpellerItem = expellerItem
	reservoir.config = config
	reservoir.run = false
	reservoir.wg = &sync.WaitGroup{}
	return reservoir, nil
}

// GetReservoir return the reservoir
func (o *Reservoir) GetReservoir() ([]interface{}, error) {
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
	var prevQueue icd.Queue
	expellerQueues := make([]icd.Queue, 0)
	for i := range o.ExpellerItem.IngesterItems {
		o.wg.Add(1)
		o.ExpellerItem.IngesterItems[i].QueueItem.MonitorControl.WaitGroup = o.wg
		o.ExpellerItem.IngesterItems[i].QueueItem.Reset()
		go o.ExpellerItem.IngesterItems[i].QueueItem.Monitor()
		o.wg.Add(1)
		o.ExpellerItem.IngesterItems[i].MonitorControl.WaitGroup = o.wg
		go o.ExpellerItem.IngesterItems[i].Ingest()
		prevQueue = o.ExpellerItem.IngesterItems[i].QueueItem.Queue
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			o.wg.Add(1)
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.MonitorControl.WaitGroup = o.wg
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Reset()
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
	o.ExpellerItem.MonitorControl.DoneChan <- struct{}{}
	for i := range o.ExpellerItem.IngesterItems {
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Close()
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.MonitorControl.DoneChan <- struct{}{}
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].MonitorControl.DoneChan <- struct{}{}
		}
		o.ExpellerItem.IngesterItems[i].QueueItem.Close()
		o.ExpellerItem.IngesterItems[i].QueueItem.MonitorControl.DoneChan <- struct{}{}
		o.ExpellerItem.IngesterItems[i].MonitorControl.DoneChan <- struct{}{}
	}
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

// UpdateFinal updates stat
func (o *Reservoir) UpdateFinal() error {
	for i := range o.ExpellerItem.IngesterItems {
		o.ExpellerItem.IngesterItems[i].stats = <-o.ExpellerItem.IngesterItems[i].MonitorControl.FinalStatsChan
		o.ExpellerItem.IngesterItems[i].QueueItem.stats = <-o.ExpellerItem.IngesterItems[i].QueueItem.MonitorControl.FinalStatsChan
		for d := range o.ExpellerItem.IngesterItems[i].DigesterItems {
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].stats = <-o.ExpellerItem.IngesterItems[i].DigesterItems[d].MonitorControl.FinalStatsChan
			o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.stats = <-o.ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.MonitorControl.FinalStatsChan
		}
	}
	o.ExpellerItem.stats = <-o.ExpellerItem.MonitorControl.FinalStatsChan
	return nil
}
