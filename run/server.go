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
	reservoirs      *Reservoirs
	monitorDoneChan chan struct{}
	statsLock       sync.Mutex
	server          http.Server
}

// NewServer creates reservoirs system
func NewServer(reservoirs *Reservoirs) (*Server, error) {
	o := new(Server)
	router := httprouter.New()
	router.GET("/v1", o.GetStats)
	router.GET("/v1/stats", o.GetStats)
	router.GET("/v1/flows", o.GetFlows)
	router.GET("/v1/reservoirs", o.GetReservoirs)
	router.PUT("/v1/flows/:rname", o.StartFlow) // starts a flow
	//router.PUT("/v1/reservoirs/:rname", o.CreateReservoir) // creates a reservoir
	router.DELETE("/v1/flows/:rname", o.StopFlow) // stops a flow
	//router.DELETE("/v1/reservoirs/:rname", o.DeleteReservoir) // destroys a reservoir
	o.server = http.Server{
		Addr:    ":5514",
		Handler: router,
	}
	o.reservoirs = reservoirs
	o.monitorDoneChan = make(chan struct{})
	o.statsLock = sync.Mutex{}
	return o, nil
}

// GetStats returns process statistics
func (o *Server) GetStats(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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

// GetFlows returns the flows
func (o *Server) GetFlows(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.WithFields(log.Fields{
		"addr":     r.RemoteAddr,
		"method":   r.Method,
		"protocol": r.Proto,
		"url":      r.URL.Path,
	}).Debug("received request")

	for r := range o.reservoirs.Reservoirs {
		fmt.Fprintf(w, "%s:", r)
		for i := range o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems {
			iname := o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].Ingester.Name()
			iqname := o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Queue.Name()
			fmt.Fprintf(w, "\n  %s => %s", iname, iqname)
			for d := range o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				dname := o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].Digester.Name()
				dqname := o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue.Name()
				fmt.Fprintf(w, " => %s => %s", dname, dqname)
			}
			ename := o.reservoirs.Reservoirs[r].ExpellerItem.Expeller.Name()
			fmt.Fprintf(w, " => %s\n", ename)
		}
	}
}

// GetReservoirs returns the contents of all reservoirs
func (o *Server) GetReservoirs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.WithFields(log.Fields{
		"addr":     r.RemoteAddr,
		"method":   r.Method,
		"protocol": r.Proto,
		"url":      r.URL.Path,
	}).Debug("received request")

	o.statsLock.Lock()
	defer o.statsLock.Unlock()

	for r := range o.reservoirs.Reservoirs {
		fmt.Fprintf(w, "reservoir: %s\n", r)
		for i := range o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems {
			fmt.Fprintf(w, "ingester: %s\n", o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].monitorStats)
			fmt.Fprintf(w, "queue: %s\n", o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.monitorStats)
			for d := range o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				fmt.Fprintf(w, "digester: %s\n", o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].monitorStats)
				fmt.Fprintf(w, "queue: %s\n", o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.monitorStats)
			}
		}
		fmt.Fprintf(w, "expeller: %s\n", o.reservoirs.Reservoirs[r].ExpellerItem.monitorStats)
	}
}

// StartFlow starts a flow
func (o *Server) StartFlow(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.WithFields(log.Fields{
		"addr":     r.RemoteAddr,
		"method":   r.Method,
		"protocol": r.Proto,
		"url":      r.URL.Path,
	}).Debug("received request")

	o.statsLock.Lock()
	defer o.statsLock.Unlock()

	rname := p.ByName("rname")
	_, ok := o.reservoirs.Reservoirs[rname]
	if ok == false {
		// TODO
	} else {
		o.reservoirs.Reservoirs[rname].GoFlow(o.reservoirs.wg)
	}
}

// StopFlow stops a flow
func (o *Server) StopFlow(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.WithFields(log.Fields{
		"addr":     r.RemoteAddr,
		"method":   r.Method,
		"protocol": r.Proto,
		"url":      r.URL.Path,
	}).Debug("received request")

	o.statsLock.Lock()
	defer o.statsLock.Unlock()

	rname := p.ByName("rname")
	_, ok := o.reservoirs.Reservoirs[rname]
	if ok == false {
		// TODO
	} else {
		o.reservoirs.Reservoirs[rname].StopFlow()
	}
}

// Monitor is a thread for capturing stats
func (o *Server) Monitor(wg *sync.WaitGroup) {
	defer wg.Done()

	run := true
	for run == true {
		o.statsLock.Lock()
		for r := range o.reservoirs.Reservoirs {
			for i := range o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems {
				select {
				case stats := <-o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].monitorStatsChan:
					o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].monitorStats = stats
				default:
				}
				select {
				case stats := <-o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.monitorStatsChan:
					o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.monitorStats = stats
				default:
				}
				for d := range o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
					select {
					case stats := <-o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].monitorStatsChan:
						o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].monitorStats = stats
					default:
					}
					select {
					case stats := <-o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.monitorStatsChan:
						o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.monitorStats = stats
					default:
					}
				}
			}
			select {
			case stats := <-o.reservoirs.Reservoirs[r].ExpellerItem.monitorStatsChan:
				o.reservoirs.Reservoirs[r].ExpellerItem.monitorStats = stats
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
	s := <-sigint

	log.WithFields(log.Fields{
		"signal": s.String(),
	}).Debug("received signal")

	err := o.server.Shutdown(context.Background())
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("shutting down rest interface gracefully")
	}

	/// TODO move into senders
	for r := range o.reservoirs.Reservoirs {
		for i := range o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems {
			o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Queue.Close()
			for d := range o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue.Close()
			}
		}
	}

	// flows done
	for r := range o.reservoirs.Reservoirs {
		for i := range o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems {
			o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].flowDoneChan <- struct{}{}
			for d := range o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].flowDoneChan <- struct{}{}
			}
		}
		o.reservoirs.Reservoirs[r].ExpellerItem.flowDoneChan <- struct{}{}
	}

	// monitors done
	for r := range o.reservoirs.Reservoirs {
		for i := range o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems {
			o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.monitorDoneChan <- struct{}{}
			o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].monitorDoneChan <- struct{}{}
			for d := range o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.monitorDoneChan <- struct{}{}
				o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].monitorDoneChan <- struct{}{}
			}
		}
		o.reservoirs.Reservoirs[r].ExpellerItem.monitorDoneChan <- struct{}{}
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
