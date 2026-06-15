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
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Router                chi.Router
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
		Router:                chi.NewRouter(),
		Renderer:              views.NewRenderer(log, cfg.AssetVersion),
		DB:                    conn,
		Sessions:              sessionManager,
		CrossOriginProtection: NewCrossOriginProtection(log),
		Port:                  cfg.Port,
		Logger:                log,
	}

	s.Router.Use(s.CrossOriginProtection.Handler)
	s.Router.Use(s.Sessions.LoadAndSave)
	s.Router.Get("/public/*", app.GetPublicHandler(cfg).ServeHTTP)

	s.Router.Post("/receive-stats", s.ProcessBatch)

	s.Router.Get("/dash", s.HandleDashboard)
	s.Router.Get("/dash/example", s.HandleDashboardExample)

	s.Router.Get("/sign-in", s.HandleSignInPage)
	s.Router.Post("/sign-in", s.HandleSignIn)
	s.Router.Get("/sign-up", s.HandleSignUpPage)
	s.Router.Post("/sign-up", s.HandleSignUp)
	s.Router.Post("/logout", s.HandleLogout)

	return s
}

func (s *Server) Start() {
	addr := ":" + s.Port
	s.Logger.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, s.Router); err != nil {
		log.Fatal(err)
	}
}
