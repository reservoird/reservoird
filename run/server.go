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
	server       http.Server
	reservoirMap *ReservoirMap
	doneChan     chan struct{}
	wg           *sync.WaitGroup
}

// NewServer creates reservoirs system
func NewServer(reservoirMap *ReservoirMap) (*Server, error) {
	o := new(Server)

	// setup rest interface
	router := httprouter.New()
	router.GET("/v1", o.GetStats)
	router.GET("/v1/stats", o.GetStats) // go stats

	router.GET("/v1/flows", o.GetFlows)           // gets all flows
	router.GET("/v1/flows/:rname", o.GetFlow)     // gets a flow
	router.PUT("/v1/flows/:rname", o.StartFlow)   // starts a flow
	router.DELETE("/v1/flows/:rname", o.StopFlow) // stops a flow

	//router.GET("/v1/reservoirs", o.GetReservoirs) // gets all reservoirs
	//router.GET("/v1/reservoirs", o.GetReservoir) // gets a reservoir
	//router.PUT("/v1/reservoirs/:rname", o.CreateReservoir)     // creates a new reservoir
	//router.DELETE("/v1/reservoirs/:rname", o.DisposeReservoir) // disposes a reservoir

	o.server = http.Server{
		Addr:    ":5514",
		Handler: router,
	}

	o.reservoirMap = reservoirMap
	o.doneChan = make(chan struct{}, 1)
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

	flows := o.reservoirMap.GetFlows()

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

	rname := p.ByName("rname")
	flows := o.reservoirMap.GetFlow(rname)

	if flows == nil || len(flows) == 0 {
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

// StartFlow starts a flow
func (o *Server) StartFlow(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.WithFields(log.Fields{
		"addr":     r.RemoteAddr,
		"method":   r.Method,
		"protocol": r.Proto,
		"url":      r.URL.Path,
	}).Debug("received request")

	rname := p.ByName("rname")
	err := o.reservoirMap.Start(rname)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 page not found\n")
	} else {
		fmt.Fprintf(w, "%s: stopping flow\n", rname)
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

	rname := p.ByName("rname")
	err := o.reservoirMap.Stop(rname)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "404 page not found\n")
	} else {
		fmt.Fprintf(w, "%s: stopping flow\n", rname)
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

// StopMonitor stops monitor
func (o *Server) StopMonitor() {
	select {
	case o.doneChan <- struct{}{}:
	default:
	}
	o.wg.Wait()
}

// RunMonitor runs monitor
func (o *Server) RunMonitor() error {
	o.wg.Add(1)
	go o.Monitor()
	return nil
}

// Monitor is a thread for capturing stats
func (o *Server) Monitor() {
	defer o.wg.Done()

	run := true
	for run == true {
		o.reservoirMap.UpdateAll()

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
	o.StopMonitor()
}
