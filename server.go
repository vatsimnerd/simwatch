package simwatch

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/vatsimnerd/simwatch/config"
	"github.com/vatsimnerd/simwatch/provider"
)

type Server struct {
	provider *provider.Provider
	srv      *http.Server
	addr     string
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		provider: provider.New(cfg),
		addr:     cfg.Web.Addr,
	}
}

func (s *Server) Start() error {
	s.provider.Start()
	router := mux.NewRouter()
	router.HandleFunc("/api/updates", s.handleApiUpdates).Methods("GET")

	s.srv = &http.Server{
		Addr:    s.addr,
		Handler: router,
	}

	return s.srv.ListenAndServe()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}
