package server

import (
	"context"
	"net/http"
	"syslantern/db"
	"syslantern/views"

	"github.com/go-chi/chi/v5"
)

func (s *Server) HandleAgentLogsPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetAuthenticatedUser(r)
	agentID := chi.URLParam(r, "agentID")

	agent, ok := s.getVisibleAgent(w, r, user.TeamID, agentID)
	if !ok {
		return
	}

	logs, err := s.agentLogsData(ctx, agent.ID, user.TeamID)
	if err != nil {
		s.Logger.Warn("agent logs page: load logs", "team_id", user.TeamID, "agent_id", agentID, "err", err)
		http.Error(w, "Could not load this agent.", http.StatusInternalServerError)
		return
	}

	s.Renderer.RenderAgentLogsPage(w, views.AgentLogsPageData{
		AgentPageData: s.agentPageData(r, agent),
		Logs:          logs,
	})
}

func (s *Server) agentLogsData(ctx context.Context, agentID string, teamID int64) ([]views.AgentLogEntryData, error) {
	entries, err := s.DB.ListAgentLogEntries(ctx, db.ListAgentLogEntriesParams{
		AgentID: agentID,
		TeamID:  teamID,
		Limit:   200,
	})
	if err != nil {
		return nil, err
	}

	logs := make([]views.AgentLogEntryData, 0, len(entries))
	for _, entry := range entries {
		logs = append(logs, views.AgentLogEntryData{
			ObservedAt: entry.ObservedAt,
			Source:     entry.Source,
			Unit:       entry.Unit,
			Priority:   entry.Priority,
			Message:    entry.Message,
		})
	}
	return logs, nil
}
