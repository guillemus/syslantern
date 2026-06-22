package server

import (
	"net/http"
	"strings"
	"syslantern/db"
	"syslantern/shared"
	"syslantern/validate"
	"time"

	"github.com/bytedance/sonic"
)

func (s *Server) HandleIngest(w http.ResponseWriter, r *http.Request) {
	if !s.IsValidAgentAPIKey(r) {
		http.Error(w, "Invalid agent API key.", http.StatusUnauthorized)
		return
	}

	var payload shared.IngestEvent

	if err := validate.Unmarshal(r.Body, &payload); err != nil {
		s.Logger.Warn("ingest: parse request", "err", err)
		http.Error(w, "Send an ingest event as JSON.", http.StatusBadRequest)
		return
	}

	if err := s.DB.SaveLiveSnapshot(r.Context(), *payload.LiveSnapshot); err != nil {
		s.Logger.Error("ingest: save live snapshot", "err", err)
		http.Error(w, "Could not save ingest event.", http.StatusInternalServerError)
		return
	}

	// fixme: emit live snapshot loaded event

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) HandleConnect(w http.ResponseWriter, r *http.Request) {
	agentID := shared.AgentID(r.URL.Query().Get("agent_id"))
	agentName := r.URL.Query().Get("agent_name")
	agentVersion := r.URL.Query().Get("agent_version")

	team, ok := s.AuthenticateAgentAPIKey(r)
	if !ok {
		http.Error(w, "Invalid agent API key.", http.StatusUnauthorized)
		return
	}

	if agentID == "" {
		http.Error(w, "Send agent_id.", http.StatusBadRequest)
		return
	}

	if agentName == "" {
		agentName = string(agentID)
	}

	agent, err := s.DB.RegisterAgentForTeam(r.Context(), team.ID, string(agentID), agentName, agentVersion)
	if err != nil {
		s.Logger.Warn("connect: register agent", "err", err)
		http.Error(w, "Could not register agent.", http.StatusInternalServerError)
		return
	}
	s.AgentRegisteredBus.Emit(r.Context(), AgentRegisteredEvent{UserID: agent.UserID})

	commandsC := make(chan shared.Command, 16)
	flusher := w.(http.Flusher)

	// TODO: Are we sure we don't need more headers? are these appropiate?
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")

	ctx := r.Context()
	commandsC <- shared.Command{
		AnalyticsSnapshot: &shared.AnalyticsSnapshotCommand{
			Since: time.Now().UTC().Add(-1 * time.Hour),
		},
	}

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

func (s *Server) IsValidAgentAPIKey(r *http.Request) bool {
	_, ok := s.AuthenticateAgentAPIKey(r)
	return ok
}

func (s *Server) AuthenticateAgentAPIKey(r *http.Request) (db.Team, bool) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || token == "" {
		return db.Team{}, false
	}

	team, err := s.DB.GetTeamByAgentAPIKey(r.Context(), token)
	if err != nil {
		return team, false
	}

	return team, true
}
