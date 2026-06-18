package server

import (
	"app/shared"
	"app/views"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

func (s *Server) HandleAgentsIndex(w http.ResponseWriter, r *http.Request) {
	s.Renderer.RenderAgentsIndex(w, s.DashboardCache.List())
}

func (s *Server) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	agentID := shared.AgentID(chi.URLParam(r, "agentID"))
	data, ok := s.DashboardCache.Get(agentID)
	if !ok {
		data.AgentID = string(agentID)
	}
	s.Renderer.RenderDashboard(w, data)
}

func (s *Server) HandleDashboardEvents(w http.ResponseWriter, r *http.Request) {
	agentID := shared.AgentID(chi.URLParam(r, "agentID"))
	events := make(chan views.DashboardData, 16)
	cancel := s.DashboardBus.Subscribe(r.Context(), func(evt views.DashboardData) error {
		if evt.AgentID != string(agentID) {
			return nil
		}
		select {
		case events <- evt:
		default:
		}
		return nil
	})
	defer cancel()

	sse := datastar.NewSSE(w, r)

	data, ok := s.DashboardCache.Get(agentID)
	if ok {
		if err := sse.PatchElements(s.Renderer.RenderDashboardStatsHTML(data.Stats)); err != nil {
			s.Logger.Warn("dashboard events: patch cached stats", "err", err)
			return
		}
		if err := sse.PatchElements(s.Renderer.RenderDashboardHistoryHTML(data.Analytics)); err != nil {
			s.Logger.Warn("dashboard events: patch cached history", "err", err)
			return
		}
	}
	if !ok || !data.Analytics.HasAnalytics {
		s.CommandBus.Emit(r.Context(), shared.AgentCommand{
			AgentID: agentID,
			Command: shared.Command{
				AnalyticsSnapshot: &shared.AnalyticsSnapshotCommand{
					Since: time.Now().UTC().Add(-1 * time.Hour),
				},
			},
		})
	}

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
