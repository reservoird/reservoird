package run

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/julienschmidt/httprouter"
)

// Server struct contains what is needed to serve a rest interface
type Server struct {
	reservoirs []Reservoir
	donechans  []chan struct{}
	server     http.Server
}

// NewServer return a new server
func NewServer(reservoirs []Reservoir, donechans []chan struct{}) *Server {
	o := new(Server)
	router := httprouter.New()
	router.GET("/v1", o.Index)
	o.server = http.Server{
		Addr:    ":5514",
		Handler: router,
	}
	o.reservoirs = reservoirs
	o.donechans = donechans
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

	for d := range o.donechans {
		o.donechans[d] <- struct{}{}
	}
}

// Index return the contents of the index
func (o *Server) Index(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Fprintf(w, "index")
}

// Run runs and http server
func (o *Server) Run() error {
	go o.wait()
	err := o.server.ListenAndServe()
	if err != http.ErrServerClosed {
		return err
	}
	return nil
}
