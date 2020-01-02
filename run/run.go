package run

import (
	"fmt"
	"plugin"
	"sync"

	"github.com/reservoird/icd"
	"github.com/reservoird/reservoird/cfg"
)

// Constants used for map index
const (
	Queues    = "queues"
	Ingesters = "ingesters"
	Digesters = "digesters"
	Expellers = "expellers"
)

// QueueItem is what is needed for a queue
type QueueItem struct {
	Queue icd.Queue
}

// DigesterItem is what is needed to run a digester
type DigesterItem struct {
	QueueItem QueueItem
	Digester  icd.Digester
}

// IngesterItem is what is needed to run an ingester
type IngesterItem struct {
	QueueItem     QueueItem
	Ingester      icd.Ingester
	DigesterItems []DigesterItem
}

// ExpellerItem is what is needed to run an expeller
type ExpellerItem struct {
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
				return nil, fmt.Errorf("error New ingester function not found, expecting: New(string) (icd.Ingester, error)")
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
				return nil, fmt.Errorf("error New ingester queue function not found, expecting: New(string) (icd.Queue, error)")
			}
			queue, err := queueFunc(rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Config)
			if err != nil {
				return nil, err
			}
			queueItem := QueueItem{
				Queue: queue,
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
					return nil, fmt.Errorf("error New digester function not found, expecting: New(string) (icd.Digester, error)")
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
					return nil, fmt.Errorf("error New digester queue function not found, expecting: New(string) (icd.Queue, error)")
				}
				queue, err := queueFunc(rsv.Reservoirs[r].ExpellerItem.IngesterItems[i].Digesters[d].QueueItem.Config)
				if err != nil {
					return nil, err
				}
				queueItem := QueueItem{
					Queue: queue,
				}
				digesterItem := DigesterItem{
					QueueItem: queueItem,
					Digester:  digester,
				}
				digs = append(digs, digesterItem)
			}
			ingesterItem := IngesterItem{
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
			return nil, fmt.Errorf("error New expeller function not found, expecting New(string) (icd.Expeller, error)")
		}
		expeller, err := expellerFunc(rsv.Reservoirs[r].ExpellerItem.Config)
		if err != nil {
			return nil, err
		}
		expellerItem := ExpellerItem{
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
	statsChans := make(map[string]map[string]chan string, 0)
	clearChans := make(map[string]map[string]chan struct{}, 0)
	doneChans := make([]chan struct{}, 0)

	// start flows
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

	// start monitors
	for r := range reservoirs {
		for i := range reservoirs[r].ExpellerItem.IngesterItems {
			queueName := reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Queue.Name()
			statsChans[Queues] = make(map[string]chan string)
			statsChans[Queues][queueName] = make(chan string, 1)
			clearChans[Queues] = make(map[string]chan struct{})
			clearChans[Queues][queueName] = make(chan struct{}, 1)
			ingesterQueueDoneChan := make(chan struct{}, 1)
			doneChans = append(doneChans, ingesterQueueDoneChan)
			wg.Add(1)
			go reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Queue.Monitor(statsChans[Queues][queueName], clearChans[Queues][queueName], ingesterQueueDoneChan, wg)
			ingesterName := reservoirs[r].ExpellerItem.IngesterItems[i].Ingester.Name()
			statsChans[Ingesters] = make(map[string]chan string)
			statsChans[Ingesters][ingesterName] = make(chan string, 1)
			clearChans[Ingesters] = make(map[string]chan struct{})
			clearChans[Ingesters][ingesterName] = make(chan struct{}, 1)
			ingesterDoneChan := make(chan struct{}, 1)
			doneChans = append(doneChans, ingesterDoneChan)
			wg.Add(1)
			go reservoirs[r].ExpellerItem.IngesterItems[i].Ingester.Monitor(statsChans[Ingesters][ingesterName], clearChans[Ingesters][ingesterName], ingesterDoneChan, wg)
			for d := range reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				queueName := reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue.Name()
				statsChans[Queues] = make(map[string]chan string)
				statsChans[Queues][queueName] = make(chan string, 1)
				clearChans[Queues] = make(map[string]chan struct{})
				clearChans[Queues][queueName] = make(chan struct{}, 1)
				digesterQueueDoneChan := make(chan struct{}, 1)
				doneChans = append(doneChans, digesterQueueDoneChan)
				wg.Add(1)
				go reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue.Monitor(statsChans[Queues][queueName], clearChans[Queues][queueName], digesterQueueDoneChan, wg)
				digesterName := reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].Digester.Name()
				statsChans[Digesters] = make(map[string]chan string)
				statsChans[Digesters][digesterName] = make(chan string, 1)
				clearChans[Digesters] = make(map[string]chan struct{})
				clearChans[Digesters][digesterName] = make(chan struct{}, 1)
				digesterDoneChan := make(chan struct{}, 1)
				doneChans = append(doneChans, digesterDoneChan)
				wg.Add(1)
				go reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].Digester.Monitor(statsChans[Digesters][digesterName], clearChans[Digesters][digesterName], digesterDoneChan, wg)
			}
		}
		expellerName := reservoirs[r].ExpellerItem.Expeller.Name()
		statsChans[Expellers] = make(map[string]chan string)
		statsChans[Expellers][expellerName] = make(chan string, 1)
		clearChans[Expellers] = make(map[string]chan struct{})
		clearChans[Expellers][expellerName] = make(chan struct{}, 1)
		expellerDoneChan := make(chan struct{}, 1)
		doneChans = append(doneChans, expellerDoneChan)
		wg.Add(1)
		go reservoirs[r].ExpellerItem.Expeller.Monitor(statsChans[Expellers][expellerName], clearChans[Expellers][expellerName], expellerDoneChan, wg)
	}

	monitorDoneChan := make(chan struct{}, 1)
	doneChans = append(doneChans, monitorDoneChan)
	wg.Add(1)

	server := NewServer(reservoirs, statsChans, clearChans, doneChans)
	go server.Monitor(monitorDoneChan, wg)
	server.Serve()

	wg.Wait()
}
