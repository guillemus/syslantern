package main

import (
	"context"
	"fmt"
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
		fmt.Println(err)
		return
	}

	agent.Start(ctx)
}

func (a *Agent) Start(ctx context.Context) {
	collectMetricsTick := time.NewTicker(2 * time.Second)
	defer collectMetricsTick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-collectMetricsTick.C:
			cfg, err := a.client.GetAgentConfig(ctx, a.agent, a.host)
			if err != nil {
				fmt.Println("get agent config err:", err)
				continue
			}
			if cfg.Paused {
				fmt.Println("agent paused") // fixme: should we even log
				continue
			}
			fmt.Println("ticked, sending snapshot") // fixme: here aswell :S, maybe debug log
			if err := a.collectSaveSendLiveSnapshot(ctx); err != nil {
				fmt.Println("collect save send live snapshot err:", err)
			}
		}
	}
}

func (a *Agent) collectSaveSendLiveSnapshot(ctx context.Context) error {
	snapshot, err := collectLiveSnapshot(a.agent, a.host)
	if err != nil {
		return fmt.Errorf("collect snapshot: %w", err)
	}

	if err := a.client.SendLiveSnapshot(ctx, snapshot); err != nil {
		return fmt.Errorf("send snapshot: %w", err)
	}

	return nil
}
