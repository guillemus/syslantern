package server

import (
	"net/http"
	"syslantern/shared"
	"syslantern/views"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s *Server) HandleAgentsIndex(w http.ResponseWriter, r *http.Request) {
	// fixme: get list of agents somehow
	// s.Renderer.RenderAgentsIndex(w, s.DashboardCache.List())
}

func (s *Server) HandleDashboard(w http.ResponseWriter, r *http.Request) {
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
