package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"syslantern/shared"
	"syslantern/validate"
)

func getApiKey(r *http.Request) (string, bool) {
	apiKey, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	return apiKey, ok && apiKey != ""
}

func (s *Server) HandleIngest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	apiKey, ok := getApiKey(r)
	if !ok {
		s.Logger.Error("ingest: missing api key")
		http.Error(w, "API key not found", http.StatusUnauthorized)
		return
	}

	agent, err := s.DB.GetAgentFromAPIKey(ctx, apiKey)
	if err != nil {
		s.Logger.Error("ingest: invalid api key", "err", err)
		http.Error(w, "Invalid api key", http.StatusUnauthorized)
		return
	}

	var payload shared.IngestEvent

	if err := validate.Unmarshal(r.Body, &payload); err != nil {
		s.Logger.Error("ingest: parse request", "err", err)
		http.Error(w, "Send an ingest event as JSON.", http.StatusBadRequest)
		return
	}

	err = s.DB.SaveLiveSnapshot(ctx, agent.ID, agent.TeamID, payload.LiveSnapshot)
	if err != nil {
		s.Logger.Error("ingest: save live snapshot", "err", err)
		http.Error(w, "Could not save ingest event.", http.StatusInternalServerError)
		return
	}

	s.Logger.Debug("ingest: saved live snapshot", "team_id", agent.TeamID, "agent_id", agent.ID)

	// fixme: emit live snapshot loaded event
	s.BusSnapshotProcessed.Emit(EventSnapshotProcessed{
		TeamID:  1,
		AgentID: "",
	})

	w.WriteHeader(http.StatusNoContent)
}

const (
	ALLOW_INSTALL   = "ALLOW_INSTALL"
	DUPLICATED_HOST = "DUPLICATED_HOST"
	INVALID_API_KEY = "INVALID_API_KEY"
	DATABASE_ERROR  = "DATABASE_ERROR"
)

func (s *Server) HandleAgentAlreadyRegistered(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.Logger.Debug("agent already registered: check")
	// parse host id from json payload
	var payload struct {
		ApiKey string `json:"api_key" validate:"required"`
		HostID string `json:"host_id" validate:"required"`
	}
	if err := readBody(r, &payload); err != nil {
		writeErr(w, err, "PARSE_ERROR")
		return
	}

	// check api key
	agent, err := s.DB.GetAgentFromAPIKey(ctx, payload.ApiKey)
	if errors.Is(err, sql.ErrNoRows) {
		writeErr(w, err, INVALID_API_KEY)
		return
	}
	if err != nil {
		writeErr(w, err, DATABASE_ERROR)
		return
	}

	if !agent.HostID.Valid {
		// agen't doesn't have a host id, so it cannot be duplicated.

		// we save it
		if err := s.DB.UpdateAgentHostID(ctx, agent.ID, payload.HostID); err != nil {
			writeErr(w, err, DATABASE_ERROR)
			return
		}

		// and we allow installation
		writeText(w, ALLOW_INSTALL)
		return
	}

	// if agent.HostID is different than the payload host id, duplication might happen,
	// user should not install agent with that api key.
	if agent.HostID.String != payload.HostID {
		writeText(w, DUPLICATED_HOST)
		return
	}

	// agent host id is same, api key is valid, user can install correctly with the given api key.
	writeText(w, ALLOW_INSTALL)
}

func (s *Server) HandleAgentConfig(w http.ResponseWriter, r *http.Request) {
	// fixme: this is terrible, a get should not update
	// needs refactor

	// q := r.URL.Query()
	// agentID := q.Get("agent_id")
	// agentName := q.Get("agent_name")
	// agentVersion := q.Get("agent_version")
	//
	// team, ok := s.AuthenticateAgentAPIKey(r)
	// if !ok {
	// 	http.Error(w, "Invalid agent API key.", http.StatusUnauthorized)
	// 	return
	// }
	//
	// if agentID == "" {
	// 	http.Error(w, "Send agent_id.", http.StatusBadRequest)
	// 	return
	// }
	//
	// if agentName == "" {
	// 	agentName = string(agentID)
	// }

	// agent, err := s.DB.RegisterAgent(r.Context(), team.ID, agentID, agentName, agentVersion)
	// if err != nil {
	// 	s.Logger.Warn("agent config: register agent", "err", err)
	// 	http.Error(w, "Could not register agent.", http.StatusInternalServerError)
	// 	return
	// }
	// if agent.CreatedAt.Equal(agent.UpdatedAt) {
	// 	s.AgentRegisteredBus.Emit(r.Context(), AgentRegisteredEvent{TeamID: team.ID})
	// }
	//
	// w.Header().Set("Content-Type", "application/json")
	// config := shared.AgentConfig{Paused: agent.Status == db.AgentStatusPaused}
	// if err := sonic.ConfigDefault.NewEncoder(w).Encode(config); err != nil {
	// 	s.Logger.Warn("agent config: encode response", "err", err)
	// }
}
