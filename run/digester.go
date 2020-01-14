package run

import (
	"fmt"
	"plugin"

	"github.com/reservoird/icd"
	log "github.com/sirupsen/logrus"
)

// DigesterItem is what is needed to run a digester
type DigesterItem struct {
	QueueItem      *QueueItem
	Digester       icd.Digester
	MonitorControl *icd.MonitorControl
	stats          interface{}
}

// NewDigesterItem create a new digester
func NewDigesterItem(
	loc string,
	config string,
	queueLoc string,
	queueConfig string,
) (*DigesterItem, error) {
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
	o.MonitorControl = &icd.MonitorControl{
		StatsChan:      make(chan interface{}, 1),
		FinalStatsChan: make(chan interface{}, 1),
		ClearChan:      make(chan struct{}, 1),
		DoneChan:       make(chan struct{}, 1),
		WaitGroup:      nil,
	}
	o.stats = nil
	return o, nil
}

// Digest wraps actual call for debugging
func (o *DigesterItem) Digest(inQueue icd.Queue) {
	log.WithFields(log.Fields{
		"name": o.Digester.Name(),
		"func": "Digester.Digest(...)",
	}).Debug("=== into ===")
	o.Digester.Digest(inQueue, o.QueueItem.Queue, o.MonitorControl)
	log.WithFields(log.Fields{
		"name": o.Digester.Name(),
		"func": "Digester.Digest(...)",
	}).Debug("=== outof ===")
}
