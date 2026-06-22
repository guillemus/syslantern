package server

import (
	"context"
	"fmt"
	"net/http"
	"syslantern/shared"
	"syslantern/views"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

type AgentRegisteredEvent struct {
	UserID int64
}

func (s *Server) HandleIndexPage(w http.ResponseWriter, r *http.Request) {
	user, exists := s.GetAuthenticatedUser(w, r)
	if !exists {
		http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
		return
	}

	agents, err := s.agentsIndexData(r.Context(), user.ID)
	if err != nil {
		s.Logger.Warn("agents index: list agents", "err", err)
		http.Error(w, "Could not load your agents.", http.StatusInternalServerError)
		return
	}

	team, err := s.DB.GetTeamByID(r.Context(), user.TeamID)
	if err != nil {
		s.Logger.Warn("agents index: get team", "err", err)
		http.Error(w, "Could not load your team.", http.StatusInternalServerError)
		return
	}

	data := views.AgentsIndexPageData{
		Agents:         agents,
		InstallCommand: s.agentInstallCommand(r, team.AgentApiKey),
	}

	s.Renderer.RenderAgentsIndex(w, data)
}

func (s *Server) agentsIndexData(ctx context.Context, userID int64) ([]views.AgentsIndexData, error) {
	agents, err := s.DB.ListAgentsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	data := make([]views.AgentsIndexData, 0, len(agents))
	for _, agent := range agents {
		data = append(data, views.AgentsIndexData{
			ID:        agent.ID,
			Name:      agent.Name,
			Version:   agent.Version,
			UpdatedAt: agent.UpdatedAt,
		})
	}
	return data, nil
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
	user, exists := s.GetAuthenticatedUser(w, r)
	if !exists {
		http.Error(w, "Sign in to view your agents.", http.StatusUnauthorized)
		return
	}

	events := make(chan AgentRegisteredEvent, 16)
	cancel := s.AgentRegisteredBus.Subscribe(r.Context(), func(evt AgentRegisteredEvent) {
		if evt.UserID != user.ID {
			return
		}
		events <- evt
	})
	defer cancel()

	sse := datastar.NewSSE(w, r)
	for {
		select {
		case <-r.Context().Done():
			return
		case <-events:
			agents, err := s.agentsIndexData(r.Context(), user.ID)
			if err != nil {
				s.Logger.Warn("index events: list agents", "err", err)
				return
			}
			if err := sse.PatchElements(s.Renderer.RenderAgentsTableBodyHTML(agents)); err != nil {
				s.Logger.Warn("index events: patch table", "err", err)
				return
			}
		}
	}
}

func (s *Server) HandleAgentPage(w http.ResponseWriter, r *http.Request) {
	agentID := shared.AgentID(chi.URLParam(r, "agentID"))
	_ = agentID

	// fixme: get data somehow
	// s.Renderer.RenderDashboard(w, data)
}

func (s *Server) HandleDashboardEvents(w http.ResponseWriter, r *http.Request) {
	agentID := shared.AgentID(chi.URLParam(r, "agentID"))
	events := make(chan views.DashboardData, 16)
	cancel := s.DashboardBus.Subscribe(r.Context(), func(evt views.DashboardData) {
		if evt.AgentID != string(agentID) {
			return
		}
		events <- evt
	})
	defer cancel()

	sse := datastar.NewSSE(w, r)

	for {
		select {
		case <-r.Context().Done():
			return
		case data := <-events:
			html := s.Renderer.RenderDashboardStatsHTML(data.Stats)
			if err := sse.PatchElements(html); err != nil {
				s.Logger.Warn("dashboard events: patch stats", "err", err)
				return
			}
			if err := sse.PatchElements(s.Renderer.RenderDashboardHistoryHTML(data.Analytics)); err != nil {
				s.Logger.Warn("dashboard events: patch history", "err", err)
				return
			}
		}
	}
}
