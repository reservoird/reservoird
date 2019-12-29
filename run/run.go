package run

import (
	"fmt"
	"plugin"
	"sync"

	"github.com/reservoird/icd"
	"github.com/reservoird/reservoird/cfg"
)

// QueueItem is what is needed for a queue
type QueueItem struct {
	Config string
	Queue  icd.Queue
}

// DigesterItem is what is needed to run a digester
type DigesterItem struct {
	Config    string
	QueueItem QueueItem
	Digester  icd.Digester
}

// IngesterItem is what is needed to run an ingester
type IngesterItem struct {
	Config        string
	QueueItem     QueueItem
	Ingester      icd.Ingester
	DigesterItems []DigesterItem
}

// ExpellerItem is what is needed to run an expeller
type ExpellerItem struct {
	Config        string
	Expeller      icd.Expeller
	IngesterItems []IngesterItem
}

// Reservoir is the structure of the reservoir flow
type Reservoir struct {
	ExpellerItem ExpellerItem
}

// NewReservoirs setups the flow
func NewReservoirs(rsv cfg.Cfg) ([]Reservoir, error) {
	reservoirs := make([]Reservoir, 0)
	for r := range rsv.Reservoirs {
		ings := make([]IngesterItem, 0)
		for i := range rsv.Reservoirs[r].ExpellerItem.IngesterItems {
			ingesterPlug, err := plugin.Open(rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].Location)
			if err != nil {
				return nil, err
			}
			ingesterSymbol, err := ingesterPlug.Lookup("New")
			if err != nil {
				return nil, err
			}
			ingesterFunc, ok := ingesterSymbol.(func(string) (icd.Ingester, error))
			if ok == false {
				return nil, fmt.Errorf("error New ingester function not found")
			}
			ingester, err := ingesterFunc(rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].Config)
			if err != nil {
				return nil, err
			}
			queuePlug, err := plugin.Open(rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Location)
			if err != nil {
				return nil, err
			}
			queueSymbol, err := queuePlug.Lookup("New")
			if err != nil {
				return nil, err
			}
			queueFunc, ok := queueSymbol.(func(string) (icd.Queue, error))
			if ok == false {
				return nil, fmt.Errorf("error New ingester queue function not found")
			}
			queue, err := queueFunc(rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Config)
			if err != nil {
				return nil, err
			}
			queueItem := QueueItem{
				Config: rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Config,
				Queue:  queue,
			}
			digs := make([]DigesterItem, 0)
			for d := range rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].Digesters {
				digesterPlug, err := plugin.Open(rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].Digesters[d].Location)
				if err != nil {
					return nil, err
				}
				digesterSymbol, err := digesterPlug.Lookup("New")
				if err != nil {
					return nil, err
				}
				digesterFunc, ok := digesterSymbol.(func(string) (icd.Digester, error))
				if ok == false {
					return nil, fmt.Errorf("error New digester function not found")
				}
				digester, err := digesterFunc(rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].Digesters[d].Config)
				if err != nil {
					return nil, err
				}
				queuePlug, err := plugin.Open(rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].Digesters[d].QueueItem.Location)
				if err != nil {
					return nil, err
				}
				queueSymbol, err := queuePlug.Lookup("New")
				if err != nil {
					return nil, err
				}
				queueFunc, ok := queueSymbol.(func(string) (icd.Queue, error))
				if ok == false {
					return nil, fmt.Errorf("error New digester queue function not found")
				}
				queue, err := queueFunc(rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].Digesters[d].QueueItem.Config)
				if err != nil {
					return nil, err
				}
				queueItem := QueueItem{
					Config: rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].Digesters[d].QueueItem.Config,
					Queue:  queue,
				}
				digesterItem := DigesterItem{
					Config:    rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].Digesters[d].Config,
					QueueItem: queueItem,
					Digester:  digester,
				}
				digs = append(digs, digesterItem)
			}
			ingesterItem := IngesterItem{
				Config:        rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].Config,
				QueueItem:     queueItem,
				Ingester:      ingester,
				DigesterItems: digs,
			}
			ings = append(ings, ingesterItem)
		}
		expellerPlug, err := plugin.Open(rsv.Reservoirs[r].ExpellerItem.Location)
		if err != nil {
			return nil, err
		}
		expellerSymbol, err := expellerPlug.Lookup("New")
		if err != nil {
			return nil, err
		}
		expellerFunc, ok := expellerSymbol.(func(string) (icd.Expeller, error))
		if ok == false {
			return nil, fmt.Errorf("error New expeller function not found")
		}
		expeller, err := expellerFunc(rsv.Reservoirs[r].ExpellerItem.Config)
		if err != nil {
			return nil, err
		}
		expellerItem := ExpellerItem{
			Config:        rsv.Reservoirs[r].ExpellerItem.Config,
			Expeller:      expeller,
			IngesterItems: ings,
		}
		reservoir := Reservoir{
			ExpellerItem: expellerItem,
		}
		reservoirs = append(reservoirs, reservoir)
	}
	return reservoirs, nil
}

// Run runs the setup
func Run(reservoirs []Reservoir) {
	wg := &sync.WaitGroup{}
	doneChans := make([]chan struct{}, 0)

	for r := range reservoirs {
		var prevQueue icd.Queue
		expellerQueues := make([]icd.Queue, 0)
		for i := range reservoirs[r].ExpellerItem.IngesterItems {
			ingesterDone := make(chan struct{}, 1)
			doneChans = append(doneChans, ingesterDone)
			ingesterQueue := reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Queue
			wg.Add(1)
			go reservoirs[r].ExpellerItem.IngesterItems[i].Ingester.Ingest(ingesterQueue, ingesterDone, wg)
			prevQueue = ingesterQueue
			for d := range reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				digesterDone := make(chan struct{}, 1)
				doneChans = append(doneChans, digesterDone)
				digesterQueue := reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue
				wg.Add(1)
				go reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].Digester.Digest(prevQueue, digesterQueue, digesterDone, wg)
				prevQueue = digesterQueue
			}
			expellerQueues = append(expellerQueues, prevQueue)
		}
		expellerDone := make(chan struct{}, 1)
		doneChans = append(doneChans, expellerDone)
		wg.Add(1)
		go reservoirs[r].ExpellerItem.Expeller.Expel(expellerQueues, expellerDone, wg)

	}

	server := NewServer(reservoirs, doneChans)
	server.Run()

	wg.Wait()
}
