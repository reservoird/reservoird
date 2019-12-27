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
	ConfigFile string
	Queue      icd.Queue
}

// DigesterItem is what is needed to run a digester
type DigesterItem struct {
	ConfigFile string
	QueueItem  QueueItem
	Digester   icd.Digester
}

// IngesterItem is what is needed to run an ingester
type IngesterItem struct {
	ConfigFile    string
	QueueItem     QueueItem
	Ingester      icd.Ingester
	DigesterItems []DigesterItem
}

// ExpellerItem is what is needed to run an expeller
type ExpellerItem struct {
	ConfigFile    string
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
			ingesterSymbol, err := ingesterPlug.Lookup("Ingester")
			if err != nil {
				return nil, err
			}
			ingester, ok := ingesterSymbol.(icd.Ingester)
			if ok == false {
				return nil, fmt.Errorf("error Ingester interface not implemented")
			}
			queuePlug, err := plugin.Open(rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Location)
			if err != nil {
				return nil, err
			}
			queueSymbol, err := queuePlug.Lookup("Queue")
			if err != nil {
				return nil, err
			}
			queue, ok := queueSymbol.(icd.Queue)
			if ok == false {
				return nil, fmt.Errorf("error Queue interface not implemented")
			}
			queueItem := QueueItem{
				ConfigFile: rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.ConfigFile,
				Queue:      queue,
			}
			digs := make([]DigesterItem, 0)
			for d := range rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].Digesters {
				digesterPlug, err := plugin.Open(rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].Digesters[d].Location)
				if err != nil {
					return nil, err
				}
				digesterSymbol, err := digesterPlug.Lookup("Digester")
				if err != nil {
					return nil, err
				}
				digester, ok := digesterSymbol.(icd.Digester)
				if ok == false {
					return nil, fmt.Errorf("error Digester interface not implemented")
				}
				queuePlug, err := plugin.Open(rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Location)
				if err != nil {
					return nil, err
				}
				queueSymbol, err := queuePlug.Lookup("Queue")
				if err != nil {
					return nil, err
				}
				queue, ok := queueSymbol.(icd.Queue)
				if ok == false {
					return nil, fmt.Errorf("error Queue interface not implemented")
				}
				queueItem := QueueItem{
					ConfigFile: rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.ConfigFile,
					Queue:      queue,
				}
				digesterItem := DigesterItem{
					ConfigFile: rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].Digesters[d].ConfigFile,
					QueueItem:  queueItem,
					Digester:   digester,
				}
				digs = append(digs, digesterItem)
			}
			ingesterItem := IngesterItem{
				ConfigFile: rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].ConfigFile,
				QueueItem:  queueItem,
				Ingester:   ingester,
			}
			ings = append(ings, ingesterItem)
		}
		expellerPlug, err := plugin.Open(rsv.Reservoirs[r].ExpellerItem.Location)
		if err != nil {
			return nil, err
		}
		expellerSymbol, err := expellerPlug.Lookup("Expeller")
		if err != nil {
			return nil, err
		}
		expeller, ok := expellerSymbol.(icd.Expeller)
		if ok == false {
			return nil, fmt.Errorf("error Expeller interface not implemented")
		}
		expellerItem := ExpellerItem{
			ConfigFile: rsv.Reservoirs[r].ExpellerItem.ConfigFile,
			Expeller:   expeller,
		}
		reservoir := Reservoir{
			ExpellerItem: expellerItem,
		}
		reservoirs = append(reservoirs, reservoir)
	}
	return reservoirs, nil
}

// Cfg setups configuration
func Cfg(reservoirs []Reservoir) error {
	for r := range reservoirs {
		expellerConfig := reservoirs[r].ExpellerItem.ConfigFile
		err := reservoirs[r].ExpellerItem.Expeller.Config(expellerConfig)
		if err != nil {
			return err
		}
		for i := range reservoirs[r].ExpellerItem.IngesterItems {
			ingesterConfig := reservoirs[r].ExpellerItem.IngesterItems[i].ConfigFile
			err := reservoirs[r].ExpellerItem.IngesterItems[i].Ingester.Config(ingesterConfig)
			if err != nil {
				return err
			}
			queueConfig := reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.ConfigFile
			err = reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Queue.Config(queueConfig)
			if err != nil {
				return err
			}
			for d := range reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				digesterConfig := reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].ConfigFile
				err := reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].Digester.Config(digesterConfig)
				if err != nil {
					return err
				}
				queueConfig := reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.ConfigFile
				err = reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue.Config(queueConfig)
				if err != nil {
					return err
				}
			}
		}

	}
	return nil
}

// Run runs the setup
func Run(reservoirs []Reservoir) {
	wg := &sync.WaitGroup{}
	doneChans := make([]chan struct{}, 0)

	for r := range reservoirs {
		var prevQueue icd.Queue
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
		}
		expellerDone := make(chan struct{}, 1)
		doneChans = append(doneChans, expellerDone)
		wg.Add(1)
		go reservoirs[r].ExpellerItem.Expeller.Expel(prevQueue, expellerDone, wg)

	}

	server := NewServer(reservoirs, doneChans)
	server.Run()

	wg.Wait()
}
