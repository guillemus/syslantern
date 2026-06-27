package server

import (
	"context"
	"fmt"
	"net/http"
	"syslantern/db"
	"syslantern/views"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s *Server) HandleIndexPage(w http.ResponseWriter, r *http.Request) {
	user := s.GetAuthenticatedUser(r)

	data, err := s.indexData(r.Context(), user.TeamID)
	if err != nil {
		s.Logger.Warn("agents index: list agents", "err", err)
		http.Error(w, "Could not load your agents.", http.StatusInternalServerError)
		return
	}

	s.Renderer.RenderIndexPage(w, data)
}

func (s *Server) indexData(ctx context.Context, teamID int64) (views.AgentsIndexPageData, error) {
	agents, err := s.DB.ListAgentsForTeam(ctx, teamID)
	if err != nil {
		return views.AgentsIndexPageData{}, err
	}

	rows := make([]views.AgentRow, 0, len(agents))
	for _, agent := range agents {
		rows = append(rows, views.AgentRow{
			ID:        agent.ID,
			Name:      agent.Name,
			Version:   agent.Version,
			UpdatedAt: agent.UpdatedAt,
		})
	}
	return views.AgentsIndexPageData{Agents: rows}, nil
}

func (s *Server) agentInstallCommand(r *http.Request, agentAPIKey string) string {
	if s.Cfg.Dev {
		url := "http://host.multipass:3000"
		return fmt.Sprintf(
			"curl -fsSL %s/install.sh -o /tmp/syslantern-install.sh && chmod +x /tmp/syslantern-install.sh && sudo env SYSLANTERN_AGENT_URL=%s/public/syslantern-agent.tar.gz /tmp/syslantern-install.sh %q",
			url, url, agentAPIKey)
	}
	url := hubURL(r)
	return fmt.Sprintf("curl -fsSL %s/install.sh -o /tmp/syslantern-install.sh && chmod +x /tmp/syslantern-install.sh && sudo /tmp/syslantern-install.sh %q", url, agentAPIKey)
}

// hubURL uses the current request host so self-hosted hubs generate install
// commands that point agents back to the same URL the user is visiting.
func hubURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwardedProto := r.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		scheme = forwardedProto
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

func (s *Server) HandleIndexEvents(w http.ResponseWriter, r *http.Request) {
	user := s.GetAuthenticatedUser(r)

	agentCreatedC := s.AgentCreatedBus.Subscribe(r.Context())
	agentDeletedC := s.AgentDeletedBus.Subscribe(r.Context())

	sse := datastar.NewSSE(w, r)
	for {
		select {
		case <-r.Context().Done():
			return
		case evt := <-agentCreatedC:
			if evt.TeamID != user.TeamID {
				continue
			}

			data, err := s.indexData(r.Context(), user.TeamID)
			if err != nil {
				s.Logger.Warn("index events: list agents", "err", err)
				return
			}

			s.Renderer.PatchAgentsIndexTable(sse, data)
		case evt := <-agentDeletedC:
			if evt.TeamID != user.TeamID {
				continue
			}

			data, err := s.indexData(r.Context(), user.TeamID)
			if err != nil {
				s.Logger.Warn("index events: list agents", "err", err)
				return
			}

			s.Renderer.PatchAgentsIndexTable(sse, data)
		}
	}
}

func (s *Server) HandleAgentsNew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := s.GetAuthenticatedUser(r)

	var sig views.NewAgentDialogSignals
	if err := datastar.ReadSignals(r, &sig); err != nil {
		s.Logger.Error("agent new: read signals", "err", err)
		s.Renderer.PatchNewAgentDialogErr(w, r, "Could not read the agent details. Refresh the page and try again.")
		return
	}

	version := "unknown" // we don't know yet which version the agent has, it has not been installed
	createdAgent, err := s.DB.CreateAgentForTeam(ctx, user.TeamID, sig.NewAgentName, version)
	if err != nil {
		s.Logger.Error("agent new: create agent", "team_id", user.TeamID, "err", err)
		s.Renderer.PatchNewAgentDialogErr(w, r, "Could not add the agent. Try again.")
		return
	}

	s.AgentCreatedBus.Emit(AgentCreatedEvent{
		TeamID:  createdAgent.TeamID,
		AgentID: createdAgent.ID,
	})

	commandToInstall := s.agentInstallCommand(r, createdAgent.ApiKey)
	s.Renderer.PatchNewAgentDialogWithCopyCommmand(w, r, commandToInstall)
}

func (s *Server) HandleAgentsDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID := chi.URLParam(r, "agentID")
	user := s.GetAuthenticatedUser(r)

	err := s.DB.SetAgentStatusForTeam(ctx, db.SetAgentStatusForTeamParams{
		Status: db.AgentStatusDeleted,
		ID:     agentID,
		TeamID: user.TeamID,
	})
	if err != nil {
		s.Logger.Error("agent delete: delete agent",
			"team_id", user.TeamID,
			"agent_id", agentID,
			"err", err)
		s.Renderer.PatchDeleteAgentDialogErr(w, r)
		return
	}

	s.AgentDeletedBus.Emit(AgentDeletedEvent{
		TeamID:  user.TeamID,
		AgentID: agentID,
	})
	s.Renderer.PatchDeleteAgentDialogDeleted(w, r)
}
