package server

import (
	"app/shared"
	"app/views"
	"net/http"

	"github.com/starfederation/datastar-go/datastar"
)

func (s *Server) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	s.Renderer.RenderDashboard(w)
}

func (s *Server) HandleDashboardEvents(w http.ResponseWriter, r *http.Request) {
	events := make(chan shared.EventBatch, 16)
	cancel := s.BatchBus.Subscribe(r.Context(), func(evt shared.EventBatch) error {
		select {
		case events <- evt:
		default:
		}
		return nil
	})
	defer cancel()

	sse := datastar.NewSSE(w, r)

	// Send command to make client emit current status. This will give the user the latest state
	// of the host machine.
	s.CommandBus.Emit(r.Context(), shared.Command{})

	for {
		select {
		case <-r.Context().Done():
			return
		case batch := <-events:
			html := s.Renderer.RenderDashboardStatsHTML(views.DashboardStatsFromBatch(batch))
			if err := sse.PatchElements(html); err != nil {
				s.Logger.Warn("dashboard events: patch stats", "err", err)
				return
			}
		}
	}
}
