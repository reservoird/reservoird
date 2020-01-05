package run

import (
	"fmt"
	"plugin"

	"github.com/reservoird/icd"
)

// QueueItem is what is needed for a queue
type QueueItem struct {
	Queue            icd.Queue
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
