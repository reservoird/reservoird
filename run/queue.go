package run

import (
	"fmt"

	"github.com/reservoird/icd"
	"github.com/reservoird/proxy"

	log "github.com/sirupsen/logrus"
)

// QueueItem is what is needed for a queue
type QueueItem struct {
	Queue          icd.Queue
	MonitorControl *icd.MonitorControl
	stats          interface{}
}

// NewQueueItem creates a new queue
func NewQueueItem(
	loc string,
	config string,
	plugin proxy.Plugin,
) (*QueueItem, error) {
	plug, err := plugin.Open(loc)
	if err != nil {
		return nil, err
	}
	symbol, err := plug.Lookup("New")
	if err != nil {
		return nil, err
	}
	function, ok := symbol.(func(string) (icd.Queue, error))
	if ok == false {
		return nil, fmt.Errorf("error new queue function not found, expecting: New(string) (icd.Queue, error)")
	}
	queue, err := function(config)
	if err != nil {
		return nil, err
	}
	o := new(QueueItem)
	o.Queue = queue
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

// Reset wraps actual call for debugging
func (o *QueueItem) Reset() {
	log.WithFields(log.Fields{
		"name": o.Queue.Name(),
		"func": "Queue.Reset(...)",
	}).Debug("=== into ===")
	o.Queue.Reset()
	log.WithFields(log.Fields{
		"name": o.Queue.Name(),
		"func": "Queue.Reset(...)",
	}).Debug("=== outof ===")
}

// Close wraps actual call for debugging
func (o *QueueItem) Close() {
	log.WithFields(log.Fields{
		"name": o.Queue.Name(),
		"func": "Queue.Close(...)",
	}).Debug("=== into ===")
	o.Queue.Close()
	log.WithFields(log.Fields{
		"name": o.Queue.Name(),
		"func": "Queue.Close(...)",
	}).Debug("=== outof ===")
}

// Monitor wraps actual call for debugging
func (o *QueueItem) Monitor() {
	log.WithFields(log.Fields{
		"name": o.Queue.Name(),
		"func": "Queue.Monitor(...)",
	}).Debug("=== into ===")
	o.Queue.Monitor(o.MonitorControl)
	log.WithFields(log.Fields{
		"name": o.Queue.Name(),
		"func": "Queue.Monitor(...)",
	}).Debug("=== outof ===")
}
