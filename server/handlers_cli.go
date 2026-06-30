package server

import (
	"net/http"
	"strings"
	"syslantern/db"
	"syslantern/shared"

	"github.com/bytedance/sonic"
)

func getAPIKey(r *http.Request) (string, bool) {
	apiKey, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	return apiKey, ok && apiKey != ""
}

func (s *Server) HandleIngest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	apiKey, ok := getAPIKey(r)
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
	case db.AgentStatusCreated, db.AgentStatusResuming, db.AgentStatusRunning:
		// Created: just (re)installed. Resuming: was paused. Running: noop update.
		// In all cases we ingest what the agent sends and (re)set it to running.

		status := agent.Status

		if payload.LiveSnapshot != nil {
			newStatus, err := s.DB.SaveLiveSnapshot(ctx, agent.ID, agent.TeamID, payload.LiveSnapshot)
			if err != nil {
				s.Logger.Error("ingest: save live snapshot", "err", err)
				http.Error(w, "Could not save ingest event.", http.StatusInternalServerError)
				return
			}
			status = newStatus

			s.Logger.Debug("ingest: saved live snapshot", "team_id", agent.TeamID, "agent_id", agent.ID)

			s.BusSnapshotProcessed.Emit(EventSnapshotProcessed{
				Type:    SnapshotProcessedTypeMetrics,
				TeamID:  agent.TeamID,
				AgentID: agent.ID,
			})
		}

		if len(payload.Logs) > 0 {
			newStatus, err := s.DB.SaveLogs(ctx, agent.ID, agent.TeamID, payload.Logs)
			if err != nil {
				s.Logger.Error("ingest: save logs", "err", err)
				http.Error(w, "Could not save ingest event.", http.StatusInternalServerError)
				return
			}
			status = newStatus

			s.Logger.Debug("ingest: saved logs", "team_id", agent.TeamID, "agent_id", agent.ID, "count", len(payload.Logs))

			s.BusSnapshotProcessed.Emit(EventSnapshotProcessed{
				Type:    SnapshotProcessedTypeLogs,
				TeamID:  agent.TeamID,
				AgentID: agent.ID,
			})
		}

		writeJSON(w, shared.IngestResult{AgentStatus: status.ToShared()})
	}
}

const (
	allowInstall   = "ALLOW_INSTALL"
	duplicatedHost = "DUPLICATED_HOST"
	invalidAPIKey  = "INVALID_API_KEY"
	databaseError  = "DATABASE_ERROR"
)

// HandleAgentAlreadyRegistered checks if the given agent has been installed in another machine.
// If so, it errors out, informing the user that he needs to run the command in the original
// machine it was installed in.
func (s *Server) HandleAgentAlreadyRegistered(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	s.Logger.Debug("agent already registered: check")
	var payload struct {
		APIKey string `json:"api_key" validate:"required"`
		HostID string `json:"host_id" validate:"required"`
	}
	if err := readBody(r, &payload); err != nil {
		writeErr(w, err, "PARSE_ERROR")
		return
	}

	agent, notFound, err := s.DB.GetAgentFromAPIKey(ctx, payload.APIKey)
	if notFound {
		writeErr(w, err, invalidAPIKey)
		return
	} else if err != nil {
		writeErr(w, err, databaseError)
		return
	}

	if !agent.HostID.Valid {
		// agen't doesn't have a host id, so it cannot be duplicated.

		// we save it
		if err := s.DB.UpdateAgentHostID(ctx, agent.ID, payload.HostID); err != nil {
			writeErr(w, err, databaseError)
			return
		}

		// and we allow installation
		writeText(w, allowInstall)
		return
	}

	// if agent.HostID is different than the payload host id, duplication might happen,
	// user should not install agent with that api key.
	if agent.HostID.String != payload.HostID {
		writeText(w, duplicatedHost)
		return
	}

	// agent host id is same, api key is valid, user can install correctly with the given api key.
	writeText(w, allowInstall)
}

func (s *Server) HandleAgentConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	apiKey, ok := getAPIKey(r)
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
