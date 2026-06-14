package server

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"app"
	"app/config"
	"app/db"
	"app/logger"
	"app/views"

	"github.com/alexedwards/scs/v2"
)

type Server struct {
	Mux                   *http.ServeMux
	Renderer              *views.Renderer
	DB                    *db.Conn
	Sessions              *scs.SessionManager
	CrossOriginProtection *http.CrossOriginProtection
	Port                  string
	Logger                *slog.Logger
}

func NewServer() *Server {
	cfg := config.ParseConfig()

	conn, err := db.Connect(cfg.DBPath)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	log := logger.NewLogger(cfg)

	sessionManager := scs.New()
	sessionManager.Store = conn
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.IdleTimeout = 8 * time.Hour
	sessionManager.HashTokenInStore = true
	sessionManager.Cookie.Secure = !cfg.Dev
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode

	s := &Server{
		Mux:                   http.NewServeMux(),
		Renderer:              views.NewRenderer(log, cfg.AssetVersion),
		DB:                    conn,
		Sessions:              sessionManager,
		CrossOriginProtection: NewCrossOriginProtection(log),
		Port:                  cfg.Port,
		Logger:                log,
	}

	s.Mux.Handle("GET /public/", app.GetPublicHandler(cfg))
	s.RegisterLandingRoutes()
	s.RegisterDashboardRoutes()
	s.RegisterAuthRoutes()

	return s
}

func (s *Server) Start() {
	addr := ":" + s.Port
	s.Logger.Info("server starting", "addr", addr)
	handler := s.CrossOriginProtection.Handler(s.Sessions.LoadAndSave(s.Mux))
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
