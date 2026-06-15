package server

import (
	"net/http"
)

func (s *Server) RegisterLandingRoutes() {
	s.Router.Get("/", s.HandleLanding)
}

func (s *Server) HandleLanding(w http.ResponseWriter, r *http.Request) {
	s.Renderer.RenderLanding(w)
}
