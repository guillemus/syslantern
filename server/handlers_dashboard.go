package server

import (
	"fmt"
	"net/http"
	"syslantern/shared"
	"syslantern/views"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s *Server) HandleIndexPage(w http.ResponseWriter, r *http.Request) {
	user, exists := s.GetAuthenticatedUser(w, r)
	if !exists {
		http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
		return
	}

	agents, err := s.DB.ListAgentsForUser(r.Context(), user.ID)
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
		Agents:         make([]views.AgentsIndexData, 0, len(agents)),
		InstallCommand: agentInstallCommand(r, team.AgentApiKey),
	}
	for _, agent := range agents {
		data.Agents = append(data.Agents, views.AgentsIndexData{
			ID:        agent.ID,
			Name:      agent.Name,
			Version:   agent.Version,
			UpdatedAt: agent.UpdatedAt,
		})
	}

	s.Renderer.RenderAgentsIndex(w, data)
}

func agentInstallCommand(r *http.Request, agentAPIKey string) string {
	return fmt.Sprintf("curl -fsSL %s/install.sh | bash -s -- %q", hubURL(r), agentAPIKey)
}

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

func (s *Server) HandleAgentPage(w http.ResponseWriter, r *http.Request) {
	// fixme: get data somehow

	// agentID := shared.AgentID(chi.URLParam(r, "agentID"))
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
