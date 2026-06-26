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

	CommandBus      *EventBus[shared.AgentCommand]    // fixme: this has to go
	DashboardBus    *EventBus[views.AgentMetricsData] // fixme: this has to go
	AgentCreatedBus *EventBus[AgentCreatedEvent]
	AgentDeletedBus *EventBus[AgentDeletedEvent]
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
	sessionManager.Store = db.NewSessionStore(conn)
	sessionManager.Lifetime = 7 * 24 * time.Hour
	sessionManager.HashTokenInStore = true
	sessionManager.Cookie.Secure = !cfg.Dev
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode

	r := chi.NewRouter()
	s := &Server{
		Router:                r,
		Renderer:              views.NewRenderer(log, cfg.AssetVersion, cfg.Dev),
		DB:                    conn,
		Sessions:              sessionManager,
		CrossOriginProtection: NewCrossOriginProtection(log),
		Logger:                log,
		Cfg:                   &cfg,

		CommandBus:      NewEventBus[shared.AgentCommand](),
		DashboardBus:    NewEventBus[views.AgentMetricsData](),
		AgentCreatedBus: NewEventBus[AgentCreatedEvent](),
		AgentDeletedBus: NewEventBus[AgentDeletedEvent](),
	}

	r.Use(s.CrossOriginProtection.Handler)
	r.Use(s.Sessions.LoadAndSave)
	r.Get("/public/*", syslantern.GetPublicHandler(cfg).ServeHTTP)
	r.Get("/install.sh", s.HandleInstallScript)

	r.Get("/sign-in", s.HandleSignInPage)
	r.Post("/sign-in", s.HandleSignIn)
	r.Get("/sign-up", s.HandleSignUpPage)
	r.Post("/sign-up", s.HandleSignUp)
	r.Post("/logout", s.HandleLogout)

	r.Post("/ingest", s.HandleIngest)
	r.Get("/agent/config", s.HandleAgentConfig)

	r.Post("/agents/already-registered", s.HandleAgentAlreadyRegistered)

	if cfg.Dev {
		// useful for integration tests
		r.Get("/is-authenticated", s.HandleIsAuthenticated)
	}

	r.With(s.authMiddleware).Group(func(r chi.Router) {
		// TODO: this might have to go outside instead
		r.Get("/", s.HandleIndexPage)

		r.Get("/events", s.HandleIndexEvents)

		r.Post("/agents/new", s.HandleAgentsNew)
		r.Get("/agents/{agentID}", s.HandleAgentsPage)
		r.Post("/agents/{agentID}/delete", s.HandleAgentsDelete)
	})

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
