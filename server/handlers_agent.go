package server

import (
	"app/shared"
	"app/validate"
	"net/http"

	"github.com/bytedance/sonic"
)

func (s *Server) HandleBatch(w http.ResponseWriter, r *http.Request) {
	var payload shared.EventBatch

	if err := validate.Unmarshal(r.Body, &payload); err != nil {
		s.Logger.Warn("receive stats: parse request", "err", err)
		http.Error(w, "Send an event batch as JSON.", http.StatusBadRequest)
		return
	}

	if err := s.BatchBus.Emit(r.Context(), payload); err != nil {
		s.Logger.Warn("receive stats: emit batch", "err", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) HandleConnect(w http.ResponseWriter, r *http.Request) {
	commandsC := make(chan shared.Command, 16)
	cancel := s.CommandBus.Subscribe(r.Context(), func(evt shared.Command) error {
		select {
		case commandsC <- evt:
		default:
		}
		return nil
	})
	defer cancel()

	flusher := w.(http.Flusher)

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")

	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return

		case cmd := <-commandsC:
			b, err := sonic.Marshal(cmd)
			if err != nil {
				s.Logger.Warn("connect: encode command", "err", err)
				continue
			}

			b = append(b, '\n')
			if _, err := w.Write(b); err != nil {
				s.Logger.Warn("connect: write command", "err", err)
				return
			}

			flusher.Flush()
		}
	}
}
