package run

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// Server struct contains what is needed to serve a rest interface
type Server struct {
	reservoirs      map[string]*Reservoir
	monitorDoneChan chan struct{}
	statsLock       sync.Mutex
	server          http.Server
}

// NewServer creates reservoirs system
func NewServer(reservoirs map[string]*Reservoir) (*Server, error) {
	o := new(Server)
	router := httprouter.New()
	router.GET("/v1", o.Stats)
	router.GET("/v1/stats", o.Stats)
	router.GET("/v1/flows", o.Flows)
	router.GET("/v1/reservoirs", o.Reservoirs)
	o.server = http.Server{
		Addr:    ":5514",
		Handler: router,
	}
	o.reservoirs = reservoirs
	o.monitorDoneChan = make(chan struct{})
	o.statsLock = sync.Mutex{}
	return o, nil
}

// Stats returns process statistics
func (o *Server) Stats(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.WithFields(log.Fields{
		"addr":     r.RemoteAddr,
		"method":   r.Method,
		"protocol": r.Proto,
		"url":      r.URL.Path,
	}).Debug("received request")

	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)

	gcStats := &debug.GCStats{}
	debug.ReadGCStats(gcStats)

	buildinfo := &debug.BuildInfo{}
	buildinfo, _ = debug.ReadBuildInfo()

	rs := RuntimeStats{
		CPUs:       runtime.NumCPU(),
		Goroutines: runtime.NumGoroutine(),
		Goversion:  runtime.Version(),
		BuildInfo:  buildinfo,
		GCStats:    gcStats,
		MemStats:   memStats,
	}

	b, err := json.Marshal(rs)
	if err != nil {
		// TODO
	} else {
		fmt.Fprintf(w, "%s", string(b))
	}
}

// Flows returns the flows
func (o *Server) Flows(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.WithFields(log.Fields{
		"addr":     r.RemoteAddr,
		"method":   r.Method,
		"protocol": r.Proto,
		"url":      r.URL.Path,
	}).Debug("received request")

	for r := range o.reservoirs {
		fmt.Fprintf(w, "%s:", r)
		for i := range o.reservoirs[r].ExpellerItem.IngesterItems {
			iname := o.reservoirs[r].ExpellerItem.IngesterItems[i].Ingester.Name()
			iqname := o.reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Queue.Name()
			fmt.Fprintf(w, "\n  %s => %s", iname, iqname)
			for d := range o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				dname := o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].Digester.Name()
				dqname := o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue.Name()
				fmt.Fprintf(w, " => %s => %s", dname, dqname)
			}
			ename := o.reservoirs[r].ExpellerItem.Expeller.Name()
			fmt.Fprintf(w, " => %s\n", ename)
		}
	}
}

// Reservoirs returns the contents of all reservoirs
func (o *Server) Reservoirs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.WithFields(log.Fields{
		"addr":     r.RemoteAddr,
		"method":   r.Method,
		"protocol": r.Proto,
		"url":      r.URL.Path,
	}).Debug("received request")

	o.statsLock.Lock()
	defer o.statsLock.Unlock()

	for r := range o.reservoirs {
		fmt.Fprintf(w, "reservoir: %s\n", r)
		for i := range o.reservoirs[r].ExpellerItem.IngesterItems {
			fmt.Fprintf(w, "ingester: %s\n", o.reservoirs[r].ExpellerItem.IngesterItems[i].monitorStats)
			fmt.Fprintf(w, "queue: %s\n", o.reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.monitorStats)
			for d := range o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				fmt.Fprintf(w, "digester: %s\n", o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].monitorStats)
				fmt.Fprintf(w, "queue: %s\n", o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.monitorStats)
			}
		}
		fmt.Fprintf(w, "expeller: %s\n", o.reservoirs[r].ExpellerItem.monitorStats)
	}
}

// Monitor is a thread for capturing stats
func (o *Server) Monitor(wg *sync.WaitGroup) {
	defer wg.Done()

	run := true
	for run == true {
		o.statsLock.Lock()
		for r := range o.reservoirs {
			for i := range o.reservoirs[r].ExpellerItem.IngesterItems {
				select {
				case stats := <-o.reservoirs[r].ExpellerItem.IngesterItems[i].monitorStatsChan:
					o.reservoirs[r].ExpellerItem.IngesterItems[i].monitorStats = stats
				default:
				}
				select {
				case stats := <-o.reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.monitorStatsChan:
					o.reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.monitorStats = stats
				default:
				}
				for d := range o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
					select {
					case stats := <-o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].monitorStatsChan:
						o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].monitorStats = stats
					default:
					}
					select {
					case stats := <-o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.monitorStatsChan:
						o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.monitorStats = stats
					default:
					}
				}
			}
			select {
			case stats := <-o.reservoirs[r].ExpellerItem.monitorStatsChan:
				o.reservoirs[r].ExpellerItem.monitorStats = stats
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

// Serve runs and http server
func (o *Server) Serve() error {
	go o.wait()
	err := o.server.ListenAndServe()
	if err != http.ErrServerClosed {
		return err
	}
	return nil
}
