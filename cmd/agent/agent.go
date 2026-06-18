package main

import (
	"app/cmd/agent/db"
	"app/shared"
	"context"
	"fmt"
	"time"
)

type Agent struct {
	store  *db.Conn
	client *Client
	agent  shared.Agent
	host   shared.Host
}

func NewAgent() (*Agent, error) {
	store, err := db.Connect("data/openlogs-agent.db")
	if err != nil {
		return nil, fmt.Errorf("connect sqlite db: %w", err)
	}

	identity, host, err := collectAgentHost()
	if err != nil {
		store.Close()
		return nil, fmt.Errorf("collect agent identity: %w", err)
	}

	return &Agent{
		store:  store,
		client: NewClient(),
		agent:  identity,
		host:   host,
	}, nil
}

func StartAgent(ctx context.Context) {
	agent, err := NewAgent()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer agent.Stop()

	agent.Start(ctx)
}

func (a *Agent) Start(ctx context.Context) {
	cmdC := a.client.Connect(ctx, a.agent.ID)

	collectMetricsTick := time.NewTicker(2 * time.Second)
	defer collectMetricsTick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-collectMetricsTick.C:
			fmt.Println("ticked, sending snapshot")
			if err := a.collectSaveSendLiveSnapshot(ctx); err != nil {
				fmt.Println("collect save send live snapshot err:", err)
			}
		case cmd := <-cmdC:
			if err := a.handleCommand(ctx, cmd); err != nil {
				fmt.Println("handle command err:", err)
			}
		}
	}
}

func (a *Agent) Stop() {
	a.store.Close()
}

func (a *Agent) collectSaveSendLiveSnapshot(ctx context.Context) error {
	snapshot, err := collectLiveSnapshot(a.agent, a.host)
	if err != nil {
		return fmt.Errorf("collect snapshot: %w", err)
	}

	if err := a.store.SaveLiveSnapshot(ctx, snapshot); err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}

	if err := a.client.SendLiveSnapshot(ctx, snapshot); err != nil {
		return fmt.Errorf("send snapshot: %w", err)
	}

	return nil
}

func (a *Agent) handleCommand(ctx context.Context, cmd shared.Command) error {
	switch {
	case cmd.AnalyticsSnapshot != nil:
		return a.sendAnalyticsSnapshot(ctx, cmd.AnalyticsSnapshot.Since)
	}

	return nil
}

func (a *Agent) sendAnalyticsSnapshot(ctx context.Context, since time.Time) error {
	analytics, err := a.store.LoadAnalytics(ctx, since)
	if err != nil {
		return fmt.Errorf("load analytics: %w", err)
	}

	now := time.Now().UTC()
	analytics.ID = "analytics_" + eventID(now, 0)
	analytics.Agent = a.agent
	analytics.Host = a.host
	analytics.SentAt = now

	if err := a.client.SendAnalytics(ctx, analytics); err != nil {
		return fmt.Errorf("send analytics: %w", err)
	}

	return nil
}
