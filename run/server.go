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
	reservoirs      map[string]*Reservoir
	monitorDoneChan chan struct{}
	stats           map[string]map[int]map[string]string
	statsLock       sync.Mutex
	server          http.Server
}

// NewServer creates reservoirs system
func NewServer(reservoirs map[string]*Reservoir) (*Server, error) {
	o := new(Server)
	router := httprouter.New()
	router.GET("/v1", o.Stats)
	router.GET("/v1/s", o.Stats)
	router.GET("/v1/r", o.ReservoirsAll)
	router.GET("/v1/r/:rname", o.Reservoirs)
	o.server = http.Server{
		Addr:    ":5514",
		Handler: router,
	}
	o.reservoirs = reservoirs
	o.monitorDoneChan = make(chan struct{})
	o.stats = make(map[string]map[int]map[string]string)
	o.statsLock = sync.Mutex{}
	return o, nil
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

	/// TODO move into senders
	for r := range o.reservoirs {
		for i := range o.reservoirs[r].ExpellerItem.IngesterItems {
			o.reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Queue.Close()
			for d := range o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue.Close()
			}
		}
	}

	// flows done
	for r := range o.reservoirs {
		for i := range o.reservoirs[r].ExpellerItem.IngesterItems {
			o.reservoirs[r].ExpellerItem.IngesterItems[i].flowDoneChan <- struct{}{}
			for d := range o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].flowDoneChan <- struct{}{}
			}
		}
		o.reservoirs[r].ExpellerItem.flowDoneChan <- struct{}{}
	}

	// monitors done
	for r := range o.reservoirs {
		for i := range o.reservoirs[r].ExpellerItem.IngesterItems {
			o.reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.monitorDoneChan <- struct{}{}
			o.reservoirs[r].ExpellerItem.IngesterItems[i].monitorDoneChan <- struct{}{}
			for d := range o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.monitorDoneChan <- struct{}{}
				o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].monitorDoneChan <- struct{}{}
			}
		}
		o.reservoirs[r].ExpellerItem.monitorDoneChan <- struct{}{}
	}

	// monitor done
	o.monitorDoneChan <- struct{}{}
}

// Stats returns process statistics
func (o *Server) Stats(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(w, "stats")
}

// ReservoirsAll returns the contents of all reservoirs
func (o *Server) ReservoirsAll(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	o.statsLock.Lock()
	defer o.statsLock.Unlock()
	fmt.Fprintf(w, "reservoirs")
}

// Reservoirs returns the contents of all reservoirs
func (o *Server) Reservoirs(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	o.statsLock.Lock()
	defer o.statsLock.Unlock()

	rname := p.ByName("rname")
	_, ok := o.stats[rname]
	if ok == false {
		w.WriteHeader(http.StatusNotFound)
	} else {
		fmt.Fprintf(w, "reservoir: %s\n", rname)
		_, ok = o.stats[rname][Queues]
		if ok == true {
			fmt.Fprintf(w, "queues:\n")
			for q := range o.stats[rname][Queues] {
				fmt.Fprintf(w, "%s\n", o.stats[rname][Queues][q])
			}
		}
		_, ok = o.stats[rname][Ingesters]
		if ok == true {
			fmt.Fprintf(w, "ingesters:\n")
			for i := range o.stats[rname][Ingesters] {
				fmt.Fprintf(w, "%s\n", o.stats[rname][Ingesters][i])
			}
		}
		_, ok = o.stats[rname][Digesters]
		if ok == true {
			fmt.Fprintf(w, "digesters:\n")
			for e := range o.stats[rname][Digesters] {
				fmt.Fprintf(w, "%s\n", o.stats[rname][Digesters][e])
			}
		}
		_, ok = o.stats[rname][Expellers]
		if ok == true {
			fmt.Fprintf(w, "expellers:\n")
			for e := range o.stats[rname][Expellers] {
				fmt.Fprintf(w, "%s\n", o.stats[rname][Expellers][e])
			}
		}
	}
}

// Monitor is a thread for capturing stats
func (o *Server) Monitor(wg *sync.WaitGroup) {
	defer wg.Done()

	run := true
	for run == true {
		o.statsLock.Lock()
		for r := range o.reservoirs {
			if o.stats[r] == nil {
				o.stats[r] = make(map[int]map[string]string)
			}
			for i := range o.reservoirs[r].ExpellerItem.IngesterItems {
				if o.stats[r][Ingesters] == nil {
					o.stats[r][Ingesters] = make(map[string]string)
				}
				select {
				case stats := <-o.reservoirs[r].ExpellerItem.IngesterItems[i].monitorStatsChan:
					name := o.reservoirs[r].ExpellerItem.IngesterItems[i].Ingester.Name()
					o.stats[r][Ingesters][name] = stats
				default:
				}
				if o.stats[r][Queues] == nil {
					o.stats[r][Queues] = make(map[string]string)
				}
				select {
				case stats := <-o.reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.monitorStatsChan:
					name := o.reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Queue.Name()
					o.stats[r][Queues][name] = stats
				default:
				}
				for d := range o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
					if o.stats[r][Digesters] == nil {
						o.stats[r][Digesters] = make(map[string]string)
					}
					select {
					case stats := <-o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].monitorStatsChan:
						name := o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].Digester.Name()
						o.stats[r][Digesters][name] = stats
					default:
					}
					if o.stats[r][Queues] == nil {
						o.stats[r][Queues] = make(map[string]string)
					}
					select {
					case stats := <-o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.monitorStatsChan:
						name := o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue.Name()
						o.stats[r][Queues][name] = stats
					default:
					}
				}
			}
			if o.stats[r][Expellers] == nil {
				o.stats[r][Expellers] = make(map[string]string)
			}
			select {
			case stats := <-o.reservoirs[r].ExpellerItem.monitorStatsChan:
				name := o.reservoirs[r].ExpellerItem.Expeller.Name()
				o.stats[r][Expellers][name] = stats
			default:
			}
		}
		o.statsLock.Unlock()

		select {
		case <-o.monitorDoneChan:
			run = false
		default:
		}

		if run == true {
			time.Sleep(250 * time.Millisecond)
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
