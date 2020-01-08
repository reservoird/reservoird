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
	wg              *sync.WaitGroup
}

// NewServer creates reservoirs system
func NewServer(reservoirs *Reservoirs) (*Server, error) {
	o := new(Server)
	router := httprouter.New()
	router.GET("/v1", o.GetStats)
	router.GET("/v1/stats", o.GetStats) // go stats

	router.GET("/v1/flows", o.GetFlows)           // gets all flows
	router.GET("/v1/flows/:rname", o.GetFlow)     // gets a flow
	router.PUT("/v1/flows/:rname", o.StartFlow)   // starts a flow
	router.DELETE("/v1/flows/:rname", o.StopFlow) // stops a flow

	router.GET("/v1/reservoirs", o.GetReservoirs) // gets all reservoirs
	//router.GET("/v1/reservoirs", o.GetReservoir) // gets a reservoir
	//router.PUT("/v1/reservoirs/:rname", o.CreateReservoir) // creates a reservoir
	router.DELETE("/v1/reservoirs/:rname", o.DeleteReservoir) // destroys a reservoir

	o.server = http.Server{
		Addr:    ":5514",
		Handler: router,
	}
	o.reservoirs = reservoirs
	o.monitorDoneChan = make(chan struct{}, 1)
	o.statsLock = sync.Mutex{}
	o.wg = &sync.WaitGroup{}
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
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%v\n", err)
	} else {
		fmt.Fprintf(w, "%s\n", string(b))
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

	flows := make(map[string][]string)
	for r := range o.reservoirs.Reservoirs {
		flows[r] = make([]string, 0)
		for i := range o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems {
			flows[r] = append(flows[r], o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].Ingester.Name())
			flows[r] = append(flows[r], o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.Queue.Name())
			for d := range o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				flows[r] = append(flows[r], o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].Digester.Name())
				flows[r] = append(flows[r], o.reservoirs.Reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue.Name())
			}
		}
		flows[r] = append(flows[r], o.reservoirs.Reservoirs[r].ExpellerItem.Expeller.Name())
	}

	f := FlowStats(flows)
	b, err := json.Marshal(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%v\n", err)
	} else {
		fmt.Fprintf(w, "%s\n", string(b))
	}
}

// GetFlow returns a flow
func (o *Server) GetFlow(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.WithFields(log.Fields{
		"addr":     r.RemoteAddr,
		"method":   r.Method,
		"protocol": r.Proto,
		"url":      r.URL.Path,
	}).Debug("received request")

	rname := p.ByName("rname")
	_, ok := o.reservoirs.Reservoirs[rname]
	if ok == false {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 page not found\n")
	} else {
		flows := make(map[string][]string)
		flows[rname] = make([]string, 0)
		for i := range o.reservoirs.Reservoirs[rname].ExpellerItem.IngesterItems {
			flows[rname] = append(flows[rname], o.reservoirs.Reservoirs[rname].ExpellerItem.IngesterItems[i].Ingester.Name())
			flows[rname] = append(flows[rname], o.reservoirs.Reservoirs[rname].ExpellerItem.IngesterItems[i].QueueItem.Queue.Name())
			for d := range o.reservoirs.Reservoirs[rname].ExpellerItem.IngesterItems[i].DigesterItems {
				flows[rname] = append(flows[rname], o.reservoirs.Reservoirs[rname].ExpellerItem.IngesterItems[i].DigesterItems[d].Digester.Name())
				flows[rname] = append(flows[rname], o.reservoirs.Reservoirs[rname].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.Queue.Name())
			}
		}
		flows[rname] = append(flows[rname], o.reservoirs.Reservoirs[rname].ExpellerItem.Expeller.Name())

		f := FlowStats(flows)
		b, err := json.Marshal(f)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%v\n", err)
		} else {
			fmt.Fprintf(w, "%s\n", string(b))
		}
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
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 page not found\n")
	} else {
		o.reservoirs.Reservoirs[rname].GoFlow()
		fmt.Fprintf(w, "%s: starting flow\n", rname)
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
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 page not found\n")
	} else {
		go o.stopFlow(rname)
		fmt.Fprintf(w, "%s: stopping flow\n", rname)
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

// DeleteReservoir stops a reservoir
func (o *Server) DeleteReservoir(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
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
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 page not found\n")
	} else {
		go o.deleteReservoir(rname)
		fmt.Fprintf(w, "%s: deleting reservoir\n", rname)
	}
}

func (o *Server) stopFlow(reservoir string) {
	o.statsLock.Lock()
	o.statsLock.Unlock()
	_, ok := o.reservoirs.Reservoirs[reservoir]
	if ok == true {
		o.reservoirs.Reservoirs[reservoir].StopFlow()
		o.reservoirs.Reservoirs[reservoir].WaitFlow()
	}
}

// deleteReservoir delete a reservoir
func (o *Server) deleteReservoir(reservoir string) {
	o.statsLock.Lock()
	defer o.statsLock.Unlock()
	_, ok := o.reservoirs.Reservoirs[reservoir]
	if ok == true {
		o.reservoirs.Reservoirs[reservoir].StopFlow()
		o.reservoirs.Reservoirs[reservoir].WaitFlow()
		o.reservoirs.Reservoirs[reservoir].StopMonitor()
		o.reservoirs.Reservoirs[reservoir].WaitMonitor()
		delete(o.reservoirs.Reservoirs, reservoir)
	}
}

// updateStats updates statistics
func (o *Server) updateStats() {
	o.statsLock.Lock()
	defer o.statsLock.Unlock()
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
}

// Monitor is a thread for capturing stats
func (o *Server) Monitor() {
	defer o.wg.Done()

	run := true
	for run == true {
		o.updateStats()

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

// Run start monitoring
func (o *Server) Run() error {
	o.wg.Add(1)
	go o.Monitor()
	return nil
}

// StopMonitor stops the monitor
func (o *Server) StopMonitor() {
	select {
	case o.monitorDoneChan <- struct{}{}:
	default:
	}
}

// WaitMonitor waits for monitor to stop
func (o *Server) WaitMonitor() {
	o.wg.Wait()
}

// Cleanup stops and waits for monitor
func (o *Server) Cleanup() {
	o.StopMonitor()
	o.WaitMonitor()
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
}
