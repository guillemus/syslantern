package server

import (
	"app/validate"
	"net/http"
)

func (s *Server) ProcessBatch(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Stats string `json:"stats" validate:"required"`
	}

	if err := validate.Unmarshal(r.Body, &payload); err != nil {
		s.Logger.Warn("receive stats: parse request", "err", err)
		http.Error(w, "Send stats as JSON with a stats field.", http.StatusBadRequest)
		return
	}

	s.Logger.Info("stats received", "stats", payload.Stats)
	w.WriteHeader(http.StatusNoContent)
}
