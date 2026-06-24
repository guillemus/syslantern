package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"syslantern/db"
	"syslantern/views"

	"github.com/starfederation/datastar-go/datastar"
)

func (s *Server) HandleIndexPage(w http.ResponseWriter, r *http.Request) {
	user, exists := s.GetAuthenticatedUser(w, r)
	if !exists {
		http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
		return
	}

	agents, err := s.agentsIndexData(r.Context(), user.TeamID)
	if err != nil {
		s.Logger.Warn("agents index: list agents", "err", err)
		http.Error(w, "Could not load your agents.", http.StatusInternalServerError)
		return
	}

	s.Renderer.RenderAgentsIndex(w, views.AgentsIndexPageData{
		Agents: agents,
	})
}

func (s *Server) agentsIndexData(ctx context.Context, teamID db.TeamID) ([]views.AgentsIndexData, error) {
	agents, err := s.DB.ListAgentsForTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}

	data := make([]views.AgentsIndexData, 0, len(agents))
	for _, agent := range agents {
		data = append(data, views.AgentsIndexData{
			ID:        string(agent.ID),
			Name:      agent.Name,
			Version:   agent.Version,
			UpdatedAt: agent.UpdatedAt,
		})
	}
	return data, nil
}

func (s *Server) agentInstallCommand(r *http.Request, agentAPIKey db.AgentAPIKey) string {
	if s.Cfg.Dev {
		url := "http://host.multipass:3000"
		return fmt.Sprintf(
			"curl -fsSL %s/install.sh -o /tmp/syslantern-install.sh && chmod +x /tmp/syslantern-install.sh && sudo env SYSLANTERN_AGENT_URL=%s/public/syslantern-agent.tar.gz /tmp/syslantern-install.sh %q",
			url, url, string(agentAPIKey))
	}
	url := hubURL(r)
	return fmt.Sprintf("curl -fsSL %s/install.sh -o /tmp/syslantern-install.sh && chmod +x /tmp/syslantern-install.sh && sudo /tmp/syslantern-install.sh %q", url, string(agentAPIKey))
}

// hubURL uses the current request host so self-hosted hubs generate install
// commands that point agents back to the same URL the user is visiting.
func hubURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwardedProto := r.Header.Get("X-Forwarded-Proto"); forwardedProto != "" {
		scheme = forwardedProto
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

func (s *Server) HandleIndexEvents(w http.ResponseWriter, r *http.Request) {
	user, exists := s.GetAuthenticatedUser(w, r)
	if !exists {
		http.Error(w, "Sign in to view your agents.", http.StatusUnauthorized)
		return
	}

	agentCreatedC := s.AgentCreatedBus.Subscribe(r.Context())

	sse := datastar.NewSSE(w, r)
	for {
		select {
		case <-r.Context().Done():
			return
		case evt := <-agentCreatedC:
			if evt.TeamID != user.TeamID {
				continue
			}
			agents, err := s.agentsIndexData(r.Context(), user.TeamID)
			if err != nil {
				s.Logger.Warn("index events: list agents", "err", err)
				return
			}
			if err := sse.PatchElements(s.Renderer.RenderAgentsTableBodyHTML(agents)); err != nil {
				s.Logger.Warn("index events: patch table", "err", err)
				return
			}
		}
	}
}

func (s *Server) HandleAgentNew(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user, exists := s.GetAuthenticatedUser(w, r)
	if !exists {
		sse := datastar.NewSSE(w, r)
		sse.Redirect("/sign-in")
		return
	}

	var sig views.NewAgentDialogSignals
	if err := datastar.ReadSignals(r, &sig); err != nil {
		s.Logger.Error("agent new: read signals", "err", err)
		http.Error(w, "Could not read the agent details. Refresh the page and try again.", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(sig.NewAgentName)
	if name == "" {
		http.Error(w, "Enter an agent name.", http.StatusBadRequest)
		return
	}

	version := "unknown" // we don't know yet which version the agent has, it has not been installed
	createdAgent, err := s.DB.CreateAgentForTeam(ctx, user.TeamID, name, version)
	if err != nil {
		s.Logger.Error("agent new: create agent", "team_id", user.TeamID, "err", err)
		http.Error(w, "Could not add the agent. Try again.", http.StatusInternalServerError)
		return
	}

	s.AgentCreatedBus.Emit(ctx, AgentCreatedEvent{
		TeamID:  createdAgent.TeamID,
		AgentID: createdAgent.ID,
	})
}
