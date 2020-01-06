package run

import (
	"fmt"
	"plugin"

	"github.com/reservoird/icd"
)

// ExpellerItem is what is needed to run an expeller
type ExpellerItem struct {
	Expeller         icd.Expeller
	IngesterItems    []*IngesterItem
	flowDoneChan     chan struct{}
	monitorStats     string
	monitorStatsChan chan string
	monitorClearChan chan struct{}
	monitorDoneChan  chan struct{}
}

// NewExpellerItem create a new expeller
func NewExpellerItem(loc string, config string, ingesters []*IngesterItem) (*ExpellerItem, error) {
	plug, err := plugin.Open(loc)
	if err != nil {
		return nil, err
	}
	symbol, err := plug.Lookup("New")
	if err != nil {
		return nil, err
	}
	function, ok := symbol.(func(string) (icd.Expeller, error))
	if ok == false {
		return nil, fmt.Errorf("error new queue function not found, expecting: New(string) (icd.Expeller, error)")
	}
	expeller, err := function(config)
	if err != nil {
		return nil, err
	}
	o := new(ExpellerItem)
	o.Expeller = expeller
	o.IngesterItems = ingesters
	o.flowDoneChan = make(chan struct{}, 1)
	o.monitorStatsChan = make(chan string, 1)
	o.monitorClearChan = make(chan struct{}, 1)
	o.monitorDoneChan = make(chan struct{}, 1)
	return o, nil
}
