package server

import (
	"net/http"
	"strings"
	"syslantern/db"
	"syslantern/shared"
	"syslantern/validate"

	"github.com/bytedance/sonic"
)

func (s *Server) HandleIngest(w http.ResponseWriter, r *http.Request) {
	team, ok := s.AuthenticateAgentAPIKey(r)
	if !ok {
		http.Error(w, "Invalid agent API key.", http.StatusUnauthorized)
		return
	}

	var payload shared.IngestEvent

	if err := validate.Unmarshal(r.Body, &payload); err != nil {
		s.Logger.Warn("ingest: parse request", "err", err)
		http.Error(w, "Send an ingest event as JSON.", http.StatusBadRequest)
		return
	}

	if err := s.DB.SaveLiveSnapshot(r.Context(), team.ID, *payload.LiveSnapshot); err != nil {
		s.Logger.Error("ingest: save live snapshot", "err", err)
		http.Error(w, "Could not save ingest event.", http.StatusInternalServerError)
		return
	}

	// fixme: emit live snapshot loaded event

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) HandleAgentConfig(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	agentID := shared.AgentID(q.Get("agent_id"))
	agentName := q.Get("agent_name")
	agentVersion := q.Get("agent_version")

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

	// fixme: this is horrible, a get should not update
	agent, err := s.DB.RegisterAgentForTeam(r.Context(), team.ID, agentID, agentName, agentVersion)
	if err != nil {
		s.Logger.Warn("agent config: register agent", "err", err)
		http.Error(w, "Could not register agent.", http.StatusInternalServerError)
		return
	}
	if agent.CreatedAt.Equal(agent.UpdatedAt) {
		s.AgentRegisteredBus.Emit(r.Context(), AgentRegisteredEvent{TeamID: team.ID})
	}

	w.Header().Set("Content-Type", "application/json")
	config := shared.AgentConfig{Paused: agent.Paused != 0}
	if err := sonic.ConfigDefault.NewEncoder(w).Encode(config); err != nil {
		s.Logger.Warn("agent config: encode response", "err", err)
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
