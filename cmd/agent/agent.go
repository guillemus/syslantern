package main

import (
	"context"
	"fmt"
	"os"
	"syslantern/shared"
	"time"
)

type Agent struct {
	cfg    AgentConfig
	client *Client
	agent  shared.Agent
	host   shared.Host
}

func NewAgent() (*Agent, error) {
	cfg, err := ParseConfig()
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	agent, host, err := collectAgentHost()
	if err != nil {
		return nil, fmt.Errorf("collect agent identity: %w", err)
	}

	return &Agent{
		cfg:    cfg,
		client: NewClient(cfg.HubURL, cfg.AgentAPIKey),
		agent:  agent,
		host:   host,
	}, nil
}

func StartAgent(ctx context.Context) {
	agent, err := NewAgent()
	if err != nil {
		// fixme: handle err
		return
	}

	if err := agent.Start(ctx); err != nil {
		// fixme: handle err
		os.Exit(1)
	}
}

func (a *Agent) Start(ctx context.Context) error {
	// fixme: handle retry
	agentCfg, err := a.client.GetAgentConfig(ctx)
	if err != nil {
		// fixme: handle err
		return err
	}

	agentStatus := agentCfg.AgentStatus

	collectMetricsTick := time.NewTicker(2 * time.Second)
	defer collectMetricsTick.Stop()

	pollTick := time.NewTicker(5 * time.Second)
	defer pollTick.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-pollTick.C:
			if !agentStatus.ShouldAgentPoll() {
				continue
			}

			agentCfg, err := a.client.GetAgentConfig(ctx)
			if err != nil {
				// fixme: handle err
				continue
			}

			agentStatus = agentCfg.AgentStatus

		case <-collectMetricsTick.C:
			if !agentStatus.ShouldAgentSendMetrics() {
				continue
			}
			// if config is not running

			fmt.Println("ticked, sending snapshot") // fixme: here aswell :S, maybe debug log

			snapshot, err := collectLiveSnapshot(a.agent, a.host)
			if err != nil {
				// fixme: handle err
				continue
			}

			// fixme: handle retries
			result, err := a.client.SendLiveSnapshot(ctx, snapshot)
			if err != nil {
				// fixme: handle err
				continue
			}

			agentStatus = result.AgentStatus
		}
	}
}
