package simwatch

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/vatsimnerd/simwatch/config"
	"github.com/vatsimnerd/simwatch/provider"
)

type Server struct {
	provider *provider.Provider
	srv      *http.Server
	addr     string
}

var (
	log = logrus.WithField("module", "server")
)

func NewServer(cfg *config.Config) *Server {
	return &Server{
		provider: provider.New(cfg),
		addr:     cfg.Web.Addr,
	}
}

func (s *Server) Start() error {
	l := log.WithField("func", "Start")
	l.Info("starting simwatch provider")
	s.provider.Start()

	l.Info("setting up router")
	router := mux.NewRouter()
	router.HandleFunc("/api/updates", s.handleApiUpdates).Methods("GET")
	router.HandleFunc("/api/pilots", s.handleApiPilots).Methods("GET")
	router.HandleFunc("/api/pilots/{id}", s.handleApiPilotsGet).Methods("GET")

	l.WithField("addr", s.addr).Info("creating http server")
	s.srv = &http.Server{
		Addr:    s.addr,
		Handler: router,
	}

	l.Info("entering http server loop")
	return s.srv.ListenAndServe()
}

func (s *Server) Stop() error {
	l := log.WithField("func", "Stop")
	l.Info("stopping http server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}
