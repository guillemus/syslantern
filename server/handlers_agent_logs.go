package server

import (
	"context"
	"net/http"
	"syslantern/db"
	"syslantern/views"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
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
			Message:    entry.Message,
		})
	}
	return logs, nil
}

func (s *Server) HandleAgentLogsEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := GetAuthenticatedUser(r)
	agentID := chi.URLParam(r, "agentID")
	snapshotReceivedC := s.BusSnapshotProcessed.Subscribe(ctx)

	sse := datastar.NewSSE(w, r)
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-snapshotReceivedC:
			if evt.TeamID != user.TeamID || evt.AgentID != agentID {
				continue
			}
			if evt.Type != SnapshotProcessedTypeLogs {
				continue
			}

			logs, err := s.agentLogsData(ctx, agentID, user.TeamID)
			if err != nil {
				s.Logger.Warn("agent logs events: load logs", "team_id", user.TeamID, "agent_id", agentID, "err", err)
				return
			}

			s.Renderer.PatchAgentLogs(sse, logs)
		}
	}

}
