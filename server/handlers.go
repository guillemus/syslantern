package server

import (
	"app/shared"
	"app/validate"
	"net/http"
)

func (s *Server) HandleBatch(w http.ResponseWriter, r *http.Request) {
	var payload shared.EventBatch

	if err := validate.Unmarshal(r.Body, &payload); err != nil {
		s.Logger.Warn("receive stats: parse request", "err", err)
		http.Error(w, "Send stats as JSON with a stats field.", http.StatusBadRequest)
		return
	}

	s.Logger.Info("stats received", "stats", payload)
	w.WriteHeader(http.StatusNoContent)
}
