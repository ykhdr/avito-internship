package server

import (
	"github.com/gorilla/mux"
	"net/http"
	"zadanie-6105/config"
	"zadanie-6105/database"
	"zadanie-6105/store"
)

type Server struct {
	serverAddress  string
	organizationDb database.OrganizationConnector
	tenderStore    *store.TenderStore
	r              *mux.Router
}

func NewServer(cfg *config.Config, db database.OrganizationConnector) *Server {
	s := &Server{
		serverAddress:  cfg.ServerAddress,
		organizationDb: db,
		r:              mux.NewRouter().PathPrefix("/api").Subrouter(),
		tenderStore:    store.NewTendersStore(),
	}
	s.r.HandleFunc("/ping", s.ping).Methods(http.MethodGet)
	s.r.HandleFunc("/tenders", s.tenders).Methods(http.MethodGet)
	s.r.HandleFunc("/tenders/new", s.newTender).Methods(http.MethodPost)
	s.r.HandleFunc("/tenders/my", s.myTenders).Methods(http.MethodGet)
	s.r.HandleFunc("/tenders/{tenderId}/status", s.tenderStatus).Methods(http.MethodGet)
	s.r.HandleFunc("/tenders/{tenderId}/status", s.updateTenderStatus).Methods(http.MethodPut)
	s.r.HandleFunc("/tenders/{tenderId}/edit", s.editTender).Methods(http.MethodPatch)
	return s
}

func (s *Server) Router() *mux.Router {
	return s.r
}
