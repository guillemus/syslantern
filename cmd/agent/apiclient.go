package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"syslantern/shared"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	resty *resty.Client
}

func NewClient(hubURL string, agentAPIKey string) *Client {
	return &Client{
		resty: resty.New().
			SetBaseURL(hubURL).
			SetHeader("Authorization", "Bearer "+agentAPIKey),
	}
}

func (c *Client) SendLiveSnapshot(ctx context.Context, snapshot shared.LiveSnapshot) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.resty.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(shared.IngestEvent{LiveSnapshot: &snapshot}).
		Post("/ingest")
	if err != nil {
		return fmt.Errorf("send live snapshot: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf(
			"send live snapshot: %s: %s",
			resp.Status(), strings.TrimSpace(string(resp.Body())))
	}

	return nil
}

func (c *Client) GetAgentConfig(ctx context.Context, agent shared.Agent, host shared.Host) (shared.AgentConfig, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var cfg shared.AgentConfig
	resp, err := c.resty.R().
		SetContext(ctx).
		SetResult(&cfg).
		SetQueryParam("agent_id", string(agent.ID)).
		SetQueryParam("agent_name", host.Name).
		SetQueryParam("agent_version", agent.Version).
		Get("/agent/config")
	if err != nil {
		return shared.AgentConfig{}, fmt.Errorf("get agent config: %w", err)
	}

	if resp.IsError() {
		return shared.AgentConfig{}, fmt.Errorf(
			"get agent config: %s: %s",
			resp.Status(), strings.TrimSpace(string(resp.Body())))
	}

	return cfg, nil
}
