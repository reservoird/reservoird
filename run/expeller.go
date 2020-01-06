package run

import (
	"fmt"
	"plugin"
	"sync"

	"github.com/reservoird/icd"
	log "github.com/sirupsen/logrus"
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

// Expel wraps actual call for debugging
func (o *ExpellerItem) Expel(inQueues []icd.Queue, wg *sync.WaitGroup) {
	log.WithFields(log.Fields{
		"name": o.Expeller.Name(),
		"func": "Expeller.Expel(...)",
	}).Debug("=== into ===")
	o.Expeller.Expel(inQueues, o.flowDoneChan, wg)
	log.WithFields(log.Fields{
		"name": o.Expeller.Name(),
		"func": "Expeller.Expel(...)",
	}).Debug("=== outof ===")
}

// Monitor wraps actual call for debugging
func (o *ExpellerItem) Monitor(wg *sync.WaitGroup) {
	log.WithFields(log.Fields{
		"name": o.Expeller.Name(),
		"func": "Expeller.Monitor(...)",
	}).Debug("=== into ===")
	o.Expeller.Monitor(o.monitorStatsChan, o.monitorClearChan, o.monitorDoneChan, wg)
	log.WithFields(log.Fields{
		"name": o.Expeller.Name(),
		"func": "Expeller.Monitor(...)",
	}).Debug("=== outof ===")
}
