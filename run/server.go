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
func NewServer(reservoirMap *ReservoirMap, address string) (*Server, error) {
	o := new(Server)

	// setup rest interface
	router := httprouter.New()
	router.GET("/v1/stats", o.GetStats)     // go stats
	router.GET("/v1/version", o.GetVersion) // reservoird version info

	router.GET("/v1/flows", o.GetFlows)           // gets all flows
	router.GET("/v1/flows/:rname", o.GetFlow)     // gets a flow
	router.PUT("/v1/flows/:rname", o.StartFlow)   // starts a flow
	router.DELETE("/v1/flows/:rname", o.StopFlow) // stops a flow

	router.GET("/v1/reservoirs", o.GetReservoirs)              // gets all reservoirs
	router.GET("/v1/reservoirs/:rname", o.GetReservoir)        // gets a reservoir
	router.PUT("/v1/reservoirs/:rname", o.CreateReservoir)     // creates a new reservoir
	router.DELETE("/v1/reservoirs/:rname", o.DisposeReservoir) // disposes a reservoir

	o.server = http.Server{
		Addr:    address,
		Handler: router,
	}

	o.reservoirMap = reservoirMap
	o.doneChan = make(chan struct{}, 1)
	o.wg = &sync.WaitGroup{}

	return o, nil
}

func (o *Server) GetVersion(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(w, "%s\n", "hello")
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
	flow := o.reservoirMap.GetFlow(rname)

	if flow == nil || len(flow) == 0 {
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
		fmt.Fprintf(w, "%v\n", err)
	} else {
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

	rname := p.ByName("rname")
	err := o.reservoirMap.InitStop(rname)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "%v\n", err)
	} else {
		err := o.reservoirMap.UpdateFinalAndWait(rname)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "%v\n", err)
		} else {
			fmt.Fprintf(w, "%s: stopping flow\n", rname)
		}
	}
}

// GetReservoirs get reservoirs
func (o *Server) GetReservoirs(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.WithFields(log.Fields{
		"addr":     r.RemoteAddr,
		"method":   r.Method,
		"protocol": r.Proto,
		"url":      r.URL.Path,
	}).Debug("received request")

	reservoirs := o.reservoirMap.GetReservoirs()
	if reservoirs == nil || len(reservoirs) == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "no reservoirs found")
	} else {
		b, err := json.Marshal(reservoirs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%v\n", err)
		} else {
			fmt.Fprintf(w, "%s\n", string(b))
		}
	}
}

// GetReservoir get reservoir
func (o *Server) GetReservoir(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.WithFields(log.Fields{
		"addr":     r.RemoteAddr,
		"method":   r.Method,
		"protocol": r.Proto,
		"url":      r.URL.Path,
	}).Debug("received request")

	rname := p.ByName("rname")
	reservoir, stopped, disposed := o.reservoirMap.GetReservoir(rname)
	if reservoir == nil || len(reservoir) == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "%s: not found", rname)
	} else {
		reservoirs := map[string][]interface{}{
			rname:      reservoir,
			"stopped":  []interface{}{stopped},
			"disposed": []interface{}{disposed},
		}
		r := ReservoirStats(reservoirs)
		b, err := json.Marshal(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%v\n", err)
		} else {
			fmt.Fprintf(w, "%s\n", string(b))
		}
	}
}

// CreateReservoir create a reservoir
func (o *Server) CreateReservoir(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	log.WithFields(log.Fields{
		"addr":     r.RemoteAddr,
		"method":   r.Method,
		"protocol": r.Proto,
		"url":      r.URL.Path,
	}).Debug("received request")

	rname := p.ByName("rname")
	err := o.reservoirMap.Retrieve(rname)
	if err != nil {
		// create from input
	} else {
		fmt.Fprintf(w, "%s: retrieving reservoir\n", rname)
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

	rname := p.ByName("rname")
	err := o.reservoirMap.Dispose(rname)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "%s: not found\n", rname)
	} else {
		fmt.Fprintf(w, "%s: disposing reservoir\n", rname)
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
func (o *Server) RunMonitor() {
	o.wg.Add(1)
	go o.Monitor()
}

// Monitor is a thread for capturing stats
func (o *Server) Monitor() {
	log.WithFields(log.Fields{
		"func": "Server.Monitor(...)",
	}).Debug("=== into ===")
	defer o.wg.Done()

	run := true
	for run == true {
		o.reservoirMap.UpdateAll()

		select {
		case <-o.doneChan:
			run = false
		case <-time.After(time.Second):
		}
	}
	if o.reservoirMap.StoppedAll() == false {
		o.reservoirMap.UpdateFinalAll()
	}
	log.WithFields(log.Fields{
		"func": "Server.Monitor(...)",
	}).Debug("=== outof ===")
}
