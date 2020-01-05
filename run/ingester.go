package run

import (
	"fmt"
	"plugin"

	"github.com/reservoird/icd"
)

// IngesterItem is what is needed to run an ingester
type IngesterItem struct {
	QueueItem        *QueueItem
	Ingester         icd.Ingester
	DigesterItems    []*DigesterItem
	flowDoneChan     chan struct{}
	monitorStatsChan chan string
	monitorClearChan chan struct{}
	monitorDoneChan  chan struct{}
}

// NewIngesterItem creates a new ingester
func NewIngesterItem(loc string, config string, queueLoc string, queueConfig string, digesters []*DigesterItem) (*IngesterItem, error) {
	plug, err := plugin.Open(loc)
	if err != nil {
		return nil, err
	}
	symbol, err := plug.Lookup("New")
	if err != nil {
		return nil, err
	}
	function, ok := symbol.(func(string) (icd.Ingester, error))
	if ok == false {
		return nil, fmt.Errorf("error new ingester function not found, expecting: New(string) (icd.Ingester, error)")
	}
	ingester, err := function(config)
	if err != nil {
		return nil, err
	}
	queueItem, err := NewQueueItem(
		queueLoc,
		queueConfig,
	)
	if err != nil {
		return nil, err
	}
	o := new(IngesterItem)
	o.Ingester = ingester
	o.QueueItem = queueItem
	o.DigesterItems = digesters
	o.flowDoneChan = make(chan struct{}, 1)
	o.monitorStatsChan = make(chan string, 1)
	o.monitorClearChan = make(chan struct{}, 1)
	o.monitorDoneChan = make(chan struct{}, 1)
	return o, nil
}
