package server

import (
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

	data := make([]views.AgentsIndexData, 0, len(agents))
	for _, agent := range agents {
		data = append(data, views.AgentsIndexData{
			ID:        agent.ID,
			Name:      agent.Name,
			Version:   agent.Version,
			UpdatedAt: agent.UpdatedAt,
		})
	}

	s.Renderer.RenderAgentsIndex(w, data)
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
