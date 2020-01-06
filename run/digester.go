package run

import (
	"fmt"
	"plugin"
	"sync"

	"github.com/reservoird/icd"
	log "github.com/sirupsen/logrus"
)

// DigesterItem is what is needed to run a digester
type DigesterItem struct {
	QueueItem        *QueueItem
	Digester         icd.Digester
	flowDoneChan     chan struct{}
	monitorStats     string
	monitorStatsChan chan string
	monitorClearChan chan struct{}
	monitorDoneChan  chan struct{}
}

// NewDigesterItem create a new digester
func NewDigesterItem(loc string, config string, queueLoc string, queueConfig string) (*DigesterItem, error) {
	plug, err := plugin.Open(loc)
	if err != nil {
		return nil, err
	}
	symbol, err := plug.Lookup("New")
	if err != nil {
		return nil, err
	}
	function, ok := symbol.(func(string) (icd.Digester, error))
	if ok == false {
		return nil, fmt.Errorf("error new digester function not found, expecting: New(string) (icd.Digester, error)")
	}
	digester, err := function(config)
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
	o := new(DigesterItem)
	o.Digester = digester
	o.QueueItem = queueItem
	o.flowDoneChan = make(chan struct{}, 1)
	o.monitorStatsChan = make(chan string, 1)
	o.monitorClearChan = make(chan struct{}, 1)
	o.monitorDoneChan = make(chan struct{}, 1)
	return o, nil
}

// Digest wraps actual call for debugging
func (o *DigesterItem) Digest(inQueue icd.Queue, outQueue icd.Queue, wg *sync.WaitGroup) {
	log.WithFields(log.Fields{
		"name": o.Digester.Name(),
		"func": "Digester.Digest(...)",
	}).Debug("=== into ===")
	o.Digester.Digest(inQueue, outQueue, o.flowDoneChan, wg)
	log.WithFields(log.Fields{
		"name": o.Digester.Name(),
		"func": "Digester.Digest(...)",
	}).Debug("=== outof ===")
}

// Monitor wraps actual call for debugging
func (o *DigesterItem) Monitor(wg *sync.WaitGroup) {
	log.WithFields(log.Fields{
		"name": o.Digester.Name(),
		"func": "Digester.Monitor(...)",
	}).Debug("=== into ===")
	o.Digester.Monitor(o.monitorStatsChan, o.monitorClearChan, o.monitorDoneChan, wg)
	log.WithFields(log.Fields{
		"name": o.Digester.Name(),
		"func": "Digester.Monitor(...)",
	}).Debug("=== outof ===")
}
