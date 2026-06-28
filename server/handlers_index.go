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
	ctx := r.Context()
	user := GetAuthenticatedUser(r)

	data, err := s.indexData(ctx, r, user.TeamID)
	if err != nil {
		s.Logger.Warn("agents index: list agents", "err", err)
		http.Error(w, "Could not load your agents.", http.StatusInternalServerError)
		return
	}

	s.Renderer.RenderIndexPage(w, data)
}

func (s *Server) indexData(ctx context.Context, r *http.Request, teamID int64) (views.AgentsIndexPageData, error) {
	agents, err := s.DB.ListAgents(ctx, teamID)
	if err != nil {
		return views.AgentsIndexPageData{}, err
	}

	rows := make([]views.AgentRow, 0, len(agents))
	for _, agent := range agents {
		rows = append(rows, views.AgentRow{
			ID:             agent.ID,
			Name:           agent.Name,
			Version:        agent.Version,
			UpdatedAt:      agent.UpdatedAt,
			InstallCommand: s.agentInstallCommand(r, agent.ApiKey),
			Status:         string(agent.Status),
		})
	}
	return views.AgentsIndexPageData{Agents: rows}, nil
}

func (s *Server) agentInstallCommand(r *http.Request, agentAPIKey string) string {
	if s.Cfg.Dev {
		// In dev, the test agent runs inside a Multipass VM. From inside that VM,
		// localhost points to the VM itself, not to the host machine where the hub
		// serves install.sh and the agent tarball. host.multipass lets the VM reach
		// the host, so the generated install command works without extra setup.
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
	ctx := r.Context()
	user := GetAuthenticatedUser(r)

	snapshotReceivedC := s.BusSnapshotProcessed.Subscribe(ctx)
	agentCreatedC := s.BusAgentCreated.Subscribe(ctx)
	agentDeletedC := s.BusAgentDeleted.Subscribe(ctx)

	sse := datastar.NewSSE(w, r)
	for {
		select {
		case <-ctx.Done():
			return
		case evt := <-snapshotReceivedC:
			s.Logger.Debug("index events: agent created", "team_id", evt.TeamID, "agent_id", evt.AgentID)
			if evt.TeamID != user.TeamID {
				continue
			}

			data, err := s.indexData(ctx, r, user.TeamID)
			if err != nil {
				s.Logger.Warn("index events: list agents", "err", err)
				return
			}

			s.Renderer.PatchAgentsIndexTable(sse, data)

		case evt := <-agentCreatedC:
			s.Logger.Debug("index events: agent created", "team_id", evt.TeamID, "agent_id", evt.AgentID)
			if evt.TeamID != user.TeamID {
				continue
			}

			data, err := s.indexData(ctx, r, user.TeamID)
			if err != nil {
				s.Logger.Warn("index events: list agents", "err", err)
				return
			}

			s.Renderer.PatchAgentsIndexTable(sse, data)
		case evt := <-agentDeletedC:
			s.Logger.Debug("index events: agent deleted", "team_id", evt.TeamID, "agent_id", evt.AgentID)
			if evt.TeamID != user.TeamID {
				continue
			}

			data, err := s.indexData(ctx, r, user.TeamID)
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
	user := GetAuthenticatedUser(r)

	var sig views.NewAgentDialogSignals
	if err := datastar.ReadSignals(r, &sig); err != nil {
		s.Logger.Error("agent new: read signals", "err", err)
		views.PatchNewAgentDialogErr(w, r, "Could not read the agent details. Refresh the page and try again.")
		return
	}

	version := "unknown" // we don't know yet which version the agent has, it has not been installed
	createdAgent, err := s.DB.CreateAgent(ctx, user.TeamID, sig.NewAgentName, version)
	if err != nil {
		s.Logger.Error("agent new: create agent", "team_id", user.TeamID, "err", err)
		views.PatchNewAgentDialogErr(w, r, "Could not add the agent. Try again.")
		return
	}

	s.Logger.Info("agent new: created agent", "team_id", createdAgent.TeamID, "agent_id", createdAgent.ID)

	s.BusAgentCreated.Emit(EventAgentCreated{
		TeamID:  createdAgent.TeamID,
		AgentID: createdAgent.ID,
	})

	commandToInstall := s.agentInstallCommand(r, createdAgent.ApiKey)
	views.PatchNewAgentDialogWithCopyCommmand(w, r, commandToInstall)
}

func (s *Server) HandleAgentsDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID := chi.URLParam(r, "agentID")
	user := GetAuthenticatedUser(r)

	err := s.DB.DeleteAgent(ctx, db.DeleteAgentParams{
		ID:     agentID,
		TeamID: user.TeamID,
	})
	if err != nil {
		s.Logger.Error("agent delete: delete agent",
			"team_id", user.TeamID,
			"agent_id", agentID,
			"err", err)
		views.PatchDeleteAgentDialogErr(w, r)
		return
	}

	s.Logger.Info("agent delete: deleted agent", "team_id", user.TeamID, "agent_id", agentID)

	s.BusAgentDeleted.Emit(EventAgentDeleted{
		TeamID:  user.TeamID,
		AgentID: agentID,
	})
	views.PatchDeleteAgentDialogDeleted(w, r)
}
