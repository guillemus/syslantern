package server

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"app/db"
	"app/views"

	"github.com/starfederation/datastar-go/datastar"
)

type dashboardSignals struct {
	Count int `json:"count"`
}

func (s *Server) RegisterDashboardRoutes() {
}

func (s *Server) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	_, redirected := s.GetAuthenticatedUser(w, r)
	if redirected {
		return
	}

	s.Renderer.RenderDashboard(w)
}

func (s *Server) HandleDashboardExample(w http.ResponseWriter, r *http.Request) {
	_, redirected := s.GetAuthenticatedUser(w, r)
	if redirected {
		return
	}

	signals := dashboardSignals{}
	if err := datastar.ReadSignals(r, &signals); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	nextCount := signals.Count + 1
	sse := datastar.NewSSE(w, r)

	html, err := s.Renderer.RenderDashboardExampleResultHTML(views.DashboardExampleResultData{
		Count:     nextCount,
		UpdatedAt: time.Now().Format("15:04:05"),
	})
	if err != nil {
		s.Logger.Error("dashboard example: render result", "err", err)
		return
	}

	if err := sse.PatchElements(html); err != nil {
		s.Logger.Error("dashboard example: patch elements", "err", err)
		return
	}

	if err := sse.MarshalAndPatchSignals(dashboardSignals{Count: nextCount}); err != nil {
		s.Logger.Error("dashboard example: patch signals", "err", err)
	}
}

func (s *Server) GetAuthenticatedUser(w http.ResponseWriter, r *http.Request) (*db.User, bool) {
	userID := s.Sessions.GetInt64(r.Context(), "user_id")
	if userID == 0 {
		http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
		return nil, true
	}

	user, err := s.DB.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.Logger.Warn("auth: session user not found", "user_id", userID)
		} else {
			s.Logger.Error("auth: load session user", "user_id", userID, "err", err)
		}
		s.Sessions.Remove(r.Context(), "user_id")
		http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
		return nil, true
	}

	return &user, false
}
