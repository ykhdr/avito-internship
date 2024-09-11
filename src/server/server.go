package server

import (
	"github.com/gorilla/mux"
	"net/http"
	"zadanie-6105/config"
	"zadanie-6105/database"
	"zadanie-6105/store"
)

type Server struct {
	serverAddress string
	db            database.Connector
	tenderStore   *store.TenderStore
	r             *mux.Router
}

func NewServer(cfg *config.Config, db database.Connector) *Server {
	s := &Server{
		serverAddress: cfg.ServerAddress,
		db:            db,
		r:             mux.NewRouter().PathPrefix("/api").Subrouter(),
		tenderStore:   store.NewTendersStore(),
	}
	s.r.HandleFunc("/ping", s.ping).Methods(http.MethodGet)
	s.r.HandleFunc("/tenders", s.tenders).Methods(http.MethodGet)
	s.r.HandleFunc("/tenders/new", s.newTender).Methods(http.MethodPost)
	s.r.HandleFunc("/tenders/my", s.myTenders).Methods(http.MethodGet)
	s.r.HandleFunc("/tenders/{tenderId}/status", s.tenderStatus).Methods(http.MethodGet)
	s.r.HandleFunc("/tenders/{tenderId}/status", s.updateTenderStatus).Methods(http.MethodPut)
	return s
}

func (s *Server) Router() *mux.Router {
	return s.r
}
