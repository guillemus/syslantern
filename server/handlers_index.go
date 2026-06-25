package server

import (
	"context"
	"fmt"
	"net/http"
	"syslantern/db"
	"syslantern/views"

	"github.com/starfederation/datastar-go/datastar"
)

func (s *Server) HandleIndexPage(w http.ResponseWriter, r *http.Request) {
	user, exists := s.GetAuthenticatedUser(w, r)
	if !exists {
		http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
		return
	}

	data, err := s.indexData(r.Context(), user.TeamID)
	if err != nil {
		s.Logger.Warn("agents index: list agents", "err", err)
		http.Error(w, "Could not load your agents.", http.StatusInternalServerError)
		return
	}

	s.Renderer.RenderIndexPage(w, data)
}

func (s *Server) indexData(ctx context.Context, teamID db.TeamID) (views.AgentsIndexPageData, error) {
	agents, err := s.DB.ListAgentsForTeam(ctx, teamID)
	if err != nil {
		return views.AgentsIndexPageData{}, err
	}

	rows := make([]views.AgentRow, 0, len(agents))
	for _, agent := range agents {
		rows = append(rows, views.AgentRow{
			ID:        string(agent.ID),
			Name:      agent.Name,
			Version:   agent.Version,
			UpdatedAt: agent.UpdatedAt,
		})
	}
	return views.AgentsIndexPageData{Agents: rows}, nil
}

func (s *Server) agentInstallCommand(r *http.Request, agentAPIKey db.AgentAPIKey) string {
	if s.Cfg.Dev {
		url := "http://host.multipass:3000"
		return fmt.Sprintf(
			"curl -fsSL %s/install.sh -o /tmp/syslantern-install.sh && chmod +x /tmp/syslantern-install.sh && sudo env SYSLANTERN_AGENT_URL=%s/public/syslantern-agent.tar.gz /tmp/syslantern-install.sh %q",
			url, url, string(agentAPIKey))
	}
	url := hubURL(r)
	return fmt.Sprintf("curl -fsSL %s/install.sh -o /tmp/syslantern-install.sh && chmod +x /tmp/syslantern-install.sh && sudo /tmp/syslantern-install.sh %q", url, string(agentAPIKey))
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
	user, exists := s.GetAuthenticatedUser(w, r)
	if !exists {
		http.Error(w, "Sign in to view your agents.", http.StatusUnauthorized)
		return
	}

	agentCreatedC := s.AgentCreatedBus.Subscribe(r.Context())

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

			s.Renderer.PatchIndexPage(sse, data)
		}
	}
}

func (s *Server) HandleAgentNew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, exists := s.GetAuthenticatedUser(w, r)
	if !exists {
		// fixme: this should probably come from context I think
		sse := datastar.NewSSE(w, r)
		sse.Redirect("/sign-in")
		return
	}

	var sig views.NewAgentDialogSignals
	if err := datastar.ReadSignals(r, &sig); err != nil {
		s.Logger.Error("agent new: read signals", "err", err)
		s.Renderer.PatchNewAgentDialog(w, r, "Could not read the agent details. Refresh the page and try again.")
		return
	}

	version := "unknown" // we don't know yet which version the agent has, it has not been installed
	createdAgent, err := s.DB.CreateAgentForTeam(ctx, user.TeamID, sig.NewAgentName, version)
	if err != nil {
		s.Logger.Error("agent new: create agent", "team_id", user.TeamID, "err", err)
		s.Renderer.PatchNewAgentDialog(w, r, "Could not add the agent. Try again.")
		return
	}

	s.AgentCreatedBus.Emit(ctx, AgentCreatedEvent{
		TeamID:  createdAgent.TeamID,
		AgentID: createdAgent.ID,
	})
}
