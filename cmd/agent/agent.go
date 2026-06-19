package main

import (
	"context"
	"fmt"
	"syslantern/shared"
	"time"
)

type Agent struct {
	client *Client
	agent  shared.Agent
	host   shared.Host
}

func NewAgent() (*Agent, error) {
	agent, host, err := collectAgentHost()
	if err != nil {
		return nil, fmt.Errorf("collect agent identity: %w", err)
	}

	return &Agent{
		client: NewClient(),
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
		case <-cmdC:
			// TODO: handle here server commands
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
