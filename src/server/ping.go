package server

import (
	log "log/slog"
	"net/http"
)

func (s *Server) ping(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("ok"))
	if err != nil {
		log.Warn("error during write http response", "error", err)
	}
}
