package run

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
)

// Server struct contains what is needed to serve a rest interface
type Server struct {
	reservoirs []Reservoir
	statsChans map[string]chan string
	clearChans map[string]chan struct{}
	doneChans  []chan struct{}
	stats      map[string]string
	statsLock  sync.Mutex
	server     http.Server
}

// NewServer return a new server
func NewServer(reservoirs []Reservoir, statsChans map[string]chan string, clearChans map[string]chan struct{}, doneChans []chan struct{}) *Server {
	o := new(Server)
	router := httprouter.New()
	router.GET("/v1", o.Index)
	router.GET("/v1/ingesters", o.Ingesters)
	router.GET("/v1/digesters", o.Digesters)
	router.GET("/v1/expellers", o.Expellers)
	router.GET("/v1/queues", o.Queues)
	o.server = http.Server{
		Addr:    ":5514",
		Handler: router,
	}
	o.reservoirs = reservoirs
	o.statsChans = statsChans
	o.clearChans = clearChans
	o.doneChans = doneChans
	o.stats = make(map[string]string)
	o.statsLock = sync.Mutex{}
	return o
}

// wait waits until a signal to gracefully shutdown a server
func (o *Server) wait() {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sigint

	err := o.server.Shutdown(context.Background())
	if err != nil {
		fmt.Printf("error shutting down rest interface gracefully: %v\n", err)
	}

	for r := range o.reservoirs {
		for i := range o.reservoirs[r].ExpellerItem.IngesterItems {
			o.reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Queue.Close()
			for d := range o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue.Close()
			}
		}
	}

	for d := range o.doneChans {
		o.doneChans[d] <- struct{}{}
	}
}

// Index return the contents of the index
func (o *Server) Index(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(w, "index")
}

// Ingesters retuns the contents of all ingesters
func (o *Server) Ingesters(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(w, "ingesters")
}

// Digesters returns the contents of all digesters
func (o *Server) Digesters(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(w, "digesters")
}

// Expellers returns the contents of all expellers
func (o *Server) Expellers(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(w, "expellers")
}

// Queues returrn the contents of all queues
func (o *Server) Queues(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(w, "queues")
}

// Monitor is a thread for capturing stats
func (o *Server) Monitor(doneChan <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	run := true
	for run == true {
		for k := range o.statsChans {
			stats := ""
			select {
			case stats = <-o.statsChans[k]:
			default:
			}
			o.statsLock.Lock()
			o.stats[k] = stats
			o.statsLock.Unlock()
		}

		select {
		case <-doneChan:
			run = false
		default:
		}

		if run == true {
			time.Sleep(time.Millisecond)
		}
	}
}

// Serve runs and http server
func (o *Server) Serve() error {
	go o.wait()
	err := o.server.ListenAndServe()
	if err != http.ErrServerClosed {
		return err
	}
	return nil
}
