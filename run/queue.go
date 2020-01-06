package run

import (
	"fmt"
	"plugin"
	"sync"

	"github.com/reservoird/icd"

	log "github.com/sirupsen/logrus"
)

// QueueItem is what is needed for a queue
type QueueItem struct {
	Queue            icd.Queue
	monitorStats     string
	monitorStatsChan chan string
	monitorClearChan chan struct{}
	monitorDoneChan  chan struct{}
}

// NewQueueItem creates a new queue
func NewQueueItem(loc string, config string) (*QueueItem, error) {
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
	o.monitorStatsChan = make(chan string, 1)
	o.monitorClearChan = make(chan struct{}, 1)
	o.monitorDoneChan = make(chan struct{}, 1)
	return o, nil
}

// Monitor wraps actual call for debugging
func (o *QueueItem) Monitor(wg *sync.WaitGroup) {
	log.WithFields(log.Fields{
		"name": o.Queue.Name(),
		"func": "Queue.Monitor(...)",
	}).Debug("=== into ===")
	o.Queue.Monitor(o.monitorStatsChan, o.monitorClearChan, o.monitorDoneChan, wg)
	log.WithFields(log.Fields{
		"name": o.Queue.Name(),
		"func": "Queue.Monitor(...)",
	}).Debug("=== outof ===")
}
