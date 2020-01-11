package run

import (
	"fmt"
	"plugin"

	"github.com/reservoird/icd"
	log "github.com/sirupsen/logrus"
)

// ExpellerItem is what is needed to run an expeller
type ExpellerItem struct {
	Expeller      icd.Expeller
	IngesterItems []*IngesterItem
	mc            *icd.MonitorControl
	stats         interface{}
}

// NewExpellerItem create a new expeller
func NewExpellerItem(
	loc string,
	config string,
	ingesters []*IngesterItem,
) (*ExpellerItem, error) {
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
	o.mc = &icd.MonitorControl{
		StatsChan: make(chan interface{}, 1),
		ClearChan: make(chan struct{}, 1),
		DoneChan:  make(chan struct{}, 1),
		WaitGroup: nil,
	}
	o.stats = nil
	return o, nil
}

// Expel wraps actual call for debugging
func (o *ExpellerItem) Expel(inQueues []icd.Queue) {
	log.WithFields(log.Fields{
		"name": o.Expeller.Name(),
		"func": "Expeller.Expel(...)",
	}).Debug("=== into ===")
	o.Expeller.Expel(inQueues, o.mc)
	log.WithFields(log.Fields{
		"name": o.Expeller.Name(),
		"func": "Expeller.Expel(...)",
	}).Debug("=== outof ===")
}
