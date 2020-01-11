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
	"github.com/reservoird/reservoird/cfg"
	log "github.com/sirupsen/logrus"
)

// Server struct contains what is needed to serve a rest interface
type Server struct {
	server         http.Server
	reservoirs     map[string]*Reservoir
	reservoirsLock sync.Mutex
	doneChan       chan struct{}
	wg             *sync.WaitGroup
}

// NewServer creates reservoirs system
func NewServer(rsv cfg.Cfg) (*Server, error) {
	o := new(Server)

	// setup rest interface
	router := httprouter.New()
	router.GET("/v1", o.GetStats)
	router.GET("/v1/stats", o.GetStats) // go stats

	router.GET("/v1/flows", o.GetFlows)           // gets all flows
	router.GET("/v1/flows/:rname", o.GetFlow)     // gets a flow
	router.PUT("/v1/flows/:rname", o.StartFlow)   // starts a flow
	router.DELETE("/v1/flows/:rname", o.StopFlow) // stops a flow

	router.GET("/v1/reservoirs", o.GetReservoirs) // gets all reservoirs
	//router.GET("/v1/reservoirs", o.GetReservoir) // gets a reservoir
	router.PUT("/v1/reservoirs/:rname", o.CreateReservoir)     // creates a new reservoir
	router.DELETE("/v1/reservoirs/:rname", o.DisposeReservoir) // disposes a reservoir

	o.server = http.Server{
		Addr:    ":5514",
		Handler: router,
	}

	o.reservoirs = make(map[string]*Reservoir)
	for r := range rsv.Reservoirs {
		reservoir, err := NewReservoir(rsv.Reservoirs[r])
		if err != nil {
			return nil, err
		}
		o.reservoirs[reservoir.Name] = reservoir
		reservoir.Start()
	}
	o.reservoirsLock = sync.Mutex{}

	o.doneChan = make(chan struct{}, 1)
	o.wg = &sync.WaitGroup{}

	o.RunMonitor()

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

	o.reservoirsLock.Lock()
	defer o.reservoirsLock.Unlock()

	flows := make(map[string][]string)
	for r := range o.reservoirs {
		flow, err := o.reservoirs[r].GetFlow()
		if err != nil {
			continue
		}
		flows[r] = flow
	}

	if len(flows) == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 page not found\n")
	} else {
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

// GetFlow returns a flow
func (o *Server) GetFlow(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.WithFields(log.Fields{
		"addr":     r.RemoteAddr,
		"method":   r.Method,
		"protocol": r.Proto,
		"url":      r.URL.Path,
	}).Debug("received request")

	o.reservoirsLock.Lock()
	defer o.reservoirsLock.Unlock()

	rname := p.ByName("rname")
	_, ok := o.reservoirs[rname]
	if ok == false {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 page not found\n")
	} else {
		flow, err := o.reservoirs[rname].GetFlow()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "404 page not found\n")
		} else {
			flows := map[string][]string{
				rname: flow,
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

	o.reservoirsLock.Lock()
	defer o.reservoirsLock.Unlock()

	rname := p.ByName("rname")
	_, ok := o.reservoirs[rname]
	if ok == false {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 page not found\n")
	} else {
		err := o.reservoirs[rname].Start()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%v\n", err)
		} else {
			fmt.Fprintf(w, "%s: starting flow\n", rname)
		}
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

	o.reservoirsLock.Lock()
	defer o.reservoirsLock.Unlock()

	rname := p.ByName("rname")
	_, ok := o.reservoirs[rname]
	if ok == false {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 page not found\n")
	} else {
		err := o.reservoirs[rname].Stop()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%v\n", err)
		} else {
			fmt.Fprintf(w, "%s: stopping flow\n", rname)
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

	o.reservoirsLock.Lock()
	defer o.reservoirsLock.Unlock()

	for r := range o.reservoirs {
		if o.reservoirs[r].Disposed == false {
			fmt.Fprintf(w, "reservoir: %s\n", r)
			for i := range o.reservoirs[r].ExpellerItem.IngesterItems {
				fmt.Fprintf(w, "ingester: %s\n", o.reservoirs[r].ExpellerItem.IngesterItems[i].stats)
				fmt.Fprintf(w, "queue: %s\n", o.reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.stats)
				for d := range o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
					fmt.Fprintf(w, "digester: %s\n", o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].stats)
					fmt.Fprintf(w, "queue: %s\n", o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.stats)
				}
			}
			fmt.Fprintf(w, "expeller: %s\n", o.reservoirs[r].ExpellerItem.stats)
		}
	}
}

// CreateReservoir creates a reservoir
func (o *Server) CreateReservoir(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.WithFields(log.Fields{
		"addr":     r.RemoteAddr,
		"method":   r.Method,
		"protocol": r.Proto,
		"url":      r.URL.Path,
	}).Debug("received request")

	o.reservoirsLock.Lock()
	defer o.reservoirsLock.Unlock()

	rname := p.ByName("rname")
	_, ok := o.reservoirs[rname]
	if ok == false {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 page not found\n")
	} else {
		if o.reservoirs[rname].Disposed == true {
			go o.createReservoir(rname)
			fmt.Fprintf(w, "%s: creating reservoir\n", rname)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s: reservoir already created\n", rname)
		}
	}
}

// DisposeReservoir disposes a reservoir
func (o *Server) DisposeReservoir(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.WithFields(log.Fields{
		"addr":     r.RemoteAddr,
		"method":   r.Method,
		"protocol": r.Proto,
		"url":      r.URL.Path,
	}).Debug("received request")

	o.reservoirsLock.Lock()
	defer o.reservoirsLock.Unlock()

	rname := p.ByName("rname")
	_, ok := o.reservoirs[rname]
	if ok == false {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 page not found\n")
	} else {
		if o.reservoirs[rname].Disposed == true {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "%s: reservoir already disposed\n", rname)
		} else {
			go o.disposeReservoir(rname)
			fmt.Fprintf(w, "%s: deleting reservoir\n", rname)
		}
	}
}

func (o *Server) stopFlow(reservoir string) {
	o.reservoirsLock.Lock()
	o.reservoirsLock.Unlock()
	_, ok := o.reservoirs[reservoir]
	if ok == true {
		o.reservoirs[reservoir].Stop()
	}
}

// createReservoir creates a reservoir
func (o *Server) createReservoir(reservoir string) {
	o.reservoirsLock.Lock()
	defer o.reservoirsLock.Unlock()
	_, ok := o.reservoirs[reservoir]
	if ok == true && o.reservoirs[reservoir].Disposed == true {
		cfg := o.reservoirs[reservoir].config
		r, err := NewReservoir(cfg)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Error("creating new reservoir")
		}
		delete(o.reservoirs, reservoir)
		o.reservoirs[reservoir] = r
		o.reservoirs[reservoir].Start()
	} else {
		// create from config
	}
}

// disposeReservoir dispose a reservoir
func (o *Server) disposeReservoir(reservoir string) {
	o.reservoirsLock.Lock()
	defer o.reservoirsLock.Unlock()
	_, ok := o.reservoirs[reservoir]
	if ok == true && o.reservoirs[reservoir].Disposed == false {
		o.reservoirs[reservoir].Stop()
		o.reservoirs[reservoir].Disposed = true
	}
}

// cleanup waits until a signal to gracefully shutdown a server
func (o *Server) cleanup() {
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

// Serve runs and http server
func (o *Server) Serve() error {
	go o.cleanup()
	err := o.server.ListenAndServe()
	if err != http.ErrServerClosed {
		return err
	}
	return nil
}

// initStopMonitor initiates stop of monitor
func (o *Server) initStopMonitor() {
	select {
	case o.doneChan <- struct{}{}:
	default:
	}
}

// waitMonitor waits for monitor to stop
func (o *Server) waitMonitor() {
	o.wg.Wait()
}

// StopMonitor stops monitor
func (o *Server) StopMonitor() {
	o.initStopMonitor()
	o.waitMonitor()
}

// RunMonitor runs monitor
func (o *Server) RunMonitor() error {
	o.wg.Add(1)
	go o.Monitor()
	return nil
}

func (o *Server) updateStats() {
	o.reservoirsLock.Lock()
	defer o.reservoirsLock.Unlock()
	for r := range o.reservoirs {
		for i := range o.reservoirs[r].ExpellerItem.IngesterItems {
			select {
			case stats := <-o.reservoirs[r].ExpellerItem.IngesterItems[i].mc.StatsChan:
				o.reservoirs[r].ExpellerItem.IngesterItems[i].stats = stats
			default:
			}
			select {
			case stats := <-o.reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.mc.StatsChan:
				o.reservoirs[r].ExpellerItem.IngesterItems[i].QueueItem.stats = stats
			default:
			}
			for d := range o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems {
				select {
				case stats := <-o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].mc.StatsChan:
					o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].stats = stats
				default:
				}
				select {
				case stats := <-o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.mc.StatsChan:
					o.reservoirs[r].ExpellerItem.IngesterItems[i].DigesterItems[d].QueueItem.stats = stats
				default:
				}
			}
		}
		select {
		case stats := <-o.reservoirs[r].ExpellerItem.mc.StatsChan:
			o.reservoirs[r].ExpellerItem.stats = stats
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
		case <-o.doneChan:
			run = false
		default:
		}

		if run == true {
			time.Sleep(250 * time.Millisecond)
		}
	}
}

// Cleanup stop monitor
func (o *Server) Cleanup() {
	for r := range o.reservoirs {
		o.reservoirs[r].Stop()
	}
	o.StopMonitor()
}
