package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"syslantern/shared"
	"time"
)

type Agent struct {
	cfg           AgentConfig
	client        *Client
	agent         shared.Agent
	host          shared.Host
	logger        *slog.Logger
	journalCursor string
}

func NewAgent(logger *slog.Logger) (*Agent, error) {
	cfg, err := ParseConfig()
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	agent, host, err := collectAgentHost()
	if err != nil {
		return nil, fmt.Errorf("collect agent identity: %w", err)
	}

	logger = logger.With(
		"agent_version", agent.Version,
		"host_id", host.ID,
		"host_name", host.Name,
		"host_os", host.OS,
		"host_arch", host.Arch,
	)

	return &Agent{
		cfg:    cfg,
		client: NewClient(cfg.HubURL, cfg.AgentAPIKey),
		agent:  agent,
		host:   host,
		logger: logger,
	}, nil
}

func StartAgent(ctx context.Context) {
	loggerOpts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	// loggerOpts.Level = slog.LevelDebug
	logger := slog.New(slog.NewJSONHandler(os.Stdout, loggerOpts))

	agent, err := NewAgent(logger)
	if err != nil {
		logger.Error("failed to create new agent", "error", err)
		return
	}
	agent.logger.Info("agent started", "hub_url", agent.cfg.HubURL)

	if err := agent.Collect(ctx); err != nil {
		logger.Error("failed to collect metrics", "error", err)
	}
}

func (a *Agent) Collect(ctx context.Context) error {
	agentCfg, err := a.client.GetAgentConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get agent config: %w", err)
	}

	agentStatus := agentCfg.AgentStatus
	a.logger.Info("agent config loaded", "status", agentStatus)

	collectMetricsTick := time.NewTicker(10 * time.Second)
	defer collectMetricsTick.Stop()

	collectLogsTick := time.NewTicker(2 * time.Second)
	defer collectLogsTick.Stop()

	pollTick := time.NewTicker(5 * time.Second)
	defer pollTick.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-pollTick.C:
			if !agentStatus.ShouldAgentPoll() {
				a.logger.Debug("skip config poll", "status", agentStatus)
				continue
			}

			a.logger.Debug("polling config", "status", agentStatus)

			agentCfg, err := a.client.GetAgentConfig(ctx)
			if err != nil {
				return fmt.Errorf("failed to get agent config: %w", err)
			}

			oldStatus := agentStatus
			agentStatus = agentCfg.AgentStatus

			a.logger.Debug("polled config", "old_status", oldStatus, "new_status", agentStatus)

		case <-collectMetricsTick.C:
			if !agentStatus.ShouldAgentSendMetrics() {
				a.logger.Debug("skip metrics collection", "status", agentStatus)
				continue
			}

			a.logger.Debug("collecting metrics")

			snapshot, err := collectLiveSnapshot(a.agent, a.host)
			if err != nil {
				return fmt.Errorf("failed to collect snapshot: %w", err)
			}

			result, err := a.client.SendLiveSnapshot(ctx, snapshot)
			if err != nil {
				return fmt.Errorf("failed to send snapshot: %w", err)
			}

			oldStatus := agentStatus
			agentStatus = result.AgentStatus
			a.logger.Debug("sent metrics", "old_status", oldStatus, "new_status", agentStatus)

		case <-collectLogsTick.C:
			if !agentStatus.ShouldAgentSendMetrics() {
				a.logger.Debug("skip logs collection", "status", agentStatus)
				continue
			}

			a.logger.Debug("collecting logs", "cursor_set", a.journalCursor != "", "limit", 500)

			logs, nextCursor, err := a.collectJournalLogs(ctx, a.host, a.journalCursor, 500)
			if err != nil {
				return fmt.Errorf("failed to collect journal logs: %w", err)
			}
			if len(logs) == 0 {
				// seed/advance the cursor even with no new logs, otherwise the
				// first empty-cursor call (which only seeds) never persists and
				// we re-seed forever, never sending anything.
				a.journalCursor = nextCursor
				a.logger.Debug("no logs collected", "cursor_advanced", nextCursor != "")
				continue
			}

			result, err := a.client.SendLogs(ctx, logs)
			if err != nil {
				return fmt.Errorf("failed to send logs: %w", err)
			}

			a.journalCursor = nextCursor
			oldStatus := agentStatus
			agentStatus = result.AgentStatus
			a.logger.Debug("sent logs", "count", len(logs), "old_status", oldStatus, "new_status", agentStatus)
		}
	}
}
