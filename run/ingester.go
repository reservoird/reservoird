package run

import (
	"fmt"
	"plugin"

	"github.com/reservoird/icd"
	log "github.com/sirupsen/logrus"
)

// IngesterItem is what is needed to run an ingester
type IngesterItem struct {
	QueueItem      *QueueItem
	Ingester       icd.Ingester
	DigesterItems  []*DigesterItem
	MonitorControl *icd.MonitorControl
	stats          interface{}
}

// NewIngesterItem creates a new ingester
func NewIngesterItem(
	loc string,
	config string,
	queueLoc string,
	queueConfig string,
	digesters []*DigesterItem,
) (*IngesterItem, error) {
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

// Ingest wraps actual call for debugging
func (o *IngesterItem) Ingest() {
	log.WithFields(log.Fields{
		"name": o.Ingester.Name(),
		"func": "Ingester.Ingest(...)",
	}).Debug("=== into ===")
	o.Ingester.Ingest(o.QueueItem.Queue, o.MonitorControl)
	log.WithFields(log.Fields{
		"name": o.Ingester.Name(),
		"func": "Ingester.Ingest(...)",
	}).Debug("=== outof ===")
}
