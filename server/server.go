package server

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"syslantern"
	"syslantern/config"
	"syslantern/db"
	"syslantern/logger"
	"syslantern/shared"
	"syslantern/views"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Router                chi.Router
	Renderer              *views.Renderer
	DB                    *db.Conn
	Sessions              *scs.SessionManager
	CrossOriginProtection *http.CrossOriginProtection
	Cfg                   *config.Config
	Logger                *slog.Logger

	CommandBus   *EventBus[shared.AgentCommand]
	DashboardBus *EventBus[views.DashboardData]
}

func NewServer() *Server {
	return NewServerFromConfig(config.ParseConfig())
}

func NewServerFromConfig(cfg config.Config) *Server {
	conn, err := db.Connect(cfg.DBPath)
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	log := logger.NewLogger(cfg.Dev)

	sessionManager := scs.New()
	sessionManager.Store = conn
	sessionManager.Lifetime = 7 * 24 * time.Hour
	sessionManager.HashTokenInStore = true
	sessionManager.Cookie.Secure = !cfg.Dev
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode

	s := &Server{
		Router:                chi.NewRouter(),
		Renderer:              views.NewRenderer(log, cfg.AssetVersion, cfg.Dev),
		DB:                    conn,
		Sessions:              sessionManager,
		CrossOriginProtection: NewCrossOriginProtection(log),
		Logger:                log,
		Cfg:                   &cfg,

		CommandBus:   NewEventBus[shared.AgentCommand](),
		DashboardBus: NewEventBus[views.DashboardData](),
	}

	s.Router.Use(s.CrossOriginProtection.Handler)
	s.Router.Use(s.Sessions.LoadAndSave)
	s.Router.Get("/public/*", syslantern.GetPublicHandler(cfg).ServeHTTP)
	s.Router.Get("/install.sh", s.HandleInstallScript)

	s.Router.Get("/", s.HandleIndexPage)
	s.Router.Get("/sign-in", s.HandleSignInPage)
	s.Router.Post("/sign-in", s.HandleSignIn)
	s.Router.Get("/sign-up", s.HandleSignUpPage)
	s.Router.Post("/sign-up", s.HandleSignUp)
	s.Router.Post("/logout", s.HandleLogout)

	s.Router.Post("/ingest", s.HandleIngest)
	s.Router.Get("/connect", s.HandleConnect)

	s.Router.Get("/agents/{agentID}", s.HandleAgentPage)
	s.Router.Get("/agents/{agentID}/events", s.HandleDashboardEvents)

	s.Renderer.Routes = s.Router

	return s
}

func (s *Server) Start() {
	addr := ":" + s.Cfg.Port
	s.Logger.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, s.Router); err != nil {
		log.Fatal(err)
	}
}
