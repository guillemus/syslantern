package server

import (
	"app/shared"
	"app/validate"
	"net/http"

	"github.com/bytedance/sonic"
)

func (s *Server) HandleIngest(w http.ResponseWriter, r *http.Request) {
	var payload shared.IngestEvent

	if err := validate.Unmarshal(r.Body, &payload); err != nil {
		s.Logger.Warn("ingest: parse request", "err", err)
		http.Error(w, "Send an ingest event as JSON.", http.StatusBadRequest)
		return
	}

	switch {
	case payload.LiveSnapshot != nil:
		data := s.DashboardCache.UpsertLiveSnapshot(*payload.LiveSnapshot)
		if err := s.DashboardBus.Emit(r.Context(), data); err != nil {
			s.Logger.Warn("ingest: emit dashboard live snapshot", "err", err)
		}
	case payload.Analytics != nil:
		data := s.DashboardCache.UpsertAnalytics(*payload.Analytics)
		if err := s.DashboardBus.Emit(r.Context(), data); err != nil {
			s.Logger.Warn("ingest: emit dashboard analytics", "err", err)
		}
	default:
		http.Error(w, "Send a live snapshot or analytics event.", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) HandleConnect(w http.ResponseWriter, r *http.Request) {
	agentID := shared.AgentID(r.URL.Query().Get("agent_id"))
	if agentID == "" {
		http.Error(w, "Send agent_id.", http.StatusBadRequest)
		return
	}

	commandsC := make(chan shared.Command, 16)
	cancel := s.CommandBus.Subscribe(r.Context(), func(evt shared.AgentCommand) error {
		if evt.AgentID != agentID {
			return nil
		}
		select {
		case commandsC <- evt.Command:
		default:
		}
		return nil
	})
	defer cancel()

	flusher := w.(http.Flusher)

	// TODO: Are we sure we don't need more headers? are these appropiate?
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
