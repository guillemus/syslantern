package server

import (
	"net/http"
	"strings"
	"syslantern/db"
	"syslantern/shared"

	"github.com/bytedance/sonic"
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

	agent, notFound, err := s.DB.GetAgentFromAPIKey(ctx, apiKey)
	if notFound {
		s.Logger.Error("ingest: invalid api key", "err", err)
		http.Error(w, "Invalid api key", http.StatusUnauthorized)
		return
	} else if err != nil {
		s.Logger.Error("failed to get agent from api key", "err", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	var payload shared.IngestEvent
	if err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.Logger.Error("ingest: parse request", "err", err)
		http.Error(w, "parse error", http.StatusInternalServerError)
		return
	}

	switch agent.Status {
	case db.AgentStatusCreated:
		// the agent has just been installed / reinstalled on host machine, so we should set it to running

		err = s.DB.SaveLiveSnapshot(ctx, agent.ID, agent.TeamID, payload.LiveSnapshot)
		if err != nil {
			s.Logger.Error("ingest: save live snapshot", "err", err)
			http.Error(w, "Could not save ingest event.", http.StatusInternalServerError)
			return
		}

		s.Logger.Debug("ingest: saved live snapshot", "team_id", agent.TeamID, "agent_id", agent.ID)

		s.BusSnapshotProcessed.Emit(EventSnapshotProcessed{
			TeamID:  agent.TeamID,
			AgentID: agent.ID,
		})

		writeJSON(w, shared.IngestResult{
			// fixme: this is a bug! but I want to catch it in an integration test.
			// It should be 'running' always (at least as of now)
			AgentStatus: agent.Status.ToShared(),
		})
	case db.AgentStatusDeleted:
		// agent is deleted, so it should never send metrics again. The host machine can have the agent reinstalled, in which by that point it should send metrics again, but on a new agent.

		writeJSON(w, shared.IngestResult{
			AgentStatus: agent.Status.ToShared(),
		})
	case db.AgentStatusPaused:
		// agent is paused, so it should stop sending metrics until it's resumed. It will poll for status updates

		writeJSON(w, shared.IngestResult{
			AgentStatus: agent.Status.ToShared(),
		})
	case db.AgentStatusResuming:
		// agent was paused and now resuming, so we need to set the agent to running and ingest

		err = s.DB.SaveLiveSnapshot(ctx, agent.ID, agent.TeamID, payload.LiveSnapshot)
		if err != nil {
			s.Logger.Error("ingest: save live snapshot", "err", err)
			http.Error(w, "Could not save ingest event.", http.StatusInternalServerError)
			return
		}

		s.Logger.Debug("ingest: saved live snapshot", "team_id", agent.TeamID, "agent_id", agent.ID)

		s.BusSnapshotProcessed.Emit(EventSnapshotProcessed{
			TeamID:  agent.TeamID,
			AgentID: agent.ID,
		})

		writeJSON(w, shared.IngestResult{AgentStatus: agent.Status.ToShared()})
	case db.AgentStatusRunning:
		// agent status noop update. ingest

		err = s.DB.SaveLiveSnapshot(ctx, agent.ID, agent.TeamID, payload.LiveSnapshot)
		if err != nil {
			s.Logger.Error("ingest: save live snapshot", "err", err)
			http.Error(w, "Could not save ingest event.", http.StatusInternalServerError)
			return
		}

		s.Logger.Debug("ingest: saved live snapshot", "team_id", agent.TeamID, "agent_id", agent.ID)

		s.BusSnapshotProcessed.Emit(EventSnapshotProcessed{
			TeamID:  agent.TeamID,
			AgentID: agent.ID,
		})

		writeJSON(w, shared.IngestResult{AgentStatus: agent.Status.ToShared()})
	}
}

const (
	ALLOW_INSTALL   = "ALLOW_INSTALL"
	DUPLICATED_HOST = "DUPLICATED_HOST"
	INVALID_API_KEY = "INVALID_API_KEY"
	DATABASE_ERROR  = "DATABASE_ERROR"
)

// HandleAgentAlreadyRegistered checks if the given agent has been installed in another machine.
// If so, it errors out, informing the user that he needs to run the command in the original
// machine it was installed in.
func (s *Server) HandleAgentAlreadyRegistered(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.Logger.Debug("agent already registered: check")
	var payload struct {
		ApiKey string `json:"api_key" validate:"required"`
		HostID string `json:"host_id" validate:"required"`
	}
	if err := readBody(r, &payload); err != nil {
		writeErr(w, err, "PARSE_ERROR")
		return
	}

	agent, notFound, err := s.DB.GetAgentFromAPIKey(ctx, payload.ApiKey)
	if notFound {
		writeErr(w, err, INVALID_API_KEY)
		return
	} else if err != nil {
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
	ctx := r.Context()
	apiKey, ok := getApiKey(r)
	if !ok {
		s.Logger.Error("ingest: missing api key")
		http.Error(w, "API key not found", http.StatusUnauthorized)
		return
	}

	agent, notFound, err := s.DB.GetAgentFromAPIKey(ctx, apiKey)
	if notFound {
		s.Logger.Error("ingest: invalid api key", "err", err)
		http.Error(w, "Invalid api key", http.StatusUnauthorized)
		return
	} else if err != nil {
		s.Logger.Error("failed to get agent from api key", "err", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, shared.AgentConfig{
		AgentStatus: agent.Status.ToShared(),
	})
}
