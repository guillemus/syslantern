package main

import (
	"context"
	"fmt"
	"time"

	"syslantern/shared"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	resty *resty.Client
}

func NewClient(hubURL string, agentAPIKey string) *Client {
	client := resty.New().
		SetBaseURL(hubURL).
		SetTimeout(10*time.Second).
		SetHeader("Authorization", "Bearer "+agentAPIKey)

	return &Client{
		resty: client,
	}
}

func (c *Client) SendLogs(ctx context.Context, logs []shared.LogEvent) (result shared.IngestResult, err error) {
	return c.SendIngestEvent(ctx, shared.IngestEvent{Logs: logs})
}

func (c *Client) SendLiveSnapshot(
	ctx context.Context, snapshot shared.LiveSnapshot,
) (result shared.IngestResult, err error) {
	return c.SendIngestEvent(ctx, shared.IngestEvent{LiveSnapshot: &snapshot})
}

func (c *Client) SendIngestEvent(
	ctx context.Context, event shared.IngestEvent,
) (result shared.IngestResult, err error) {
	for attempt := 0; ; attempt++ {
		resp, err := c.resty.R().
			SetContext(ctx).
			SetHeader("Content-Type", "application/json").
			SetBody(event).
			SetResult(&result).
			Post("/ingest")
		if err != nil {
			if ctx.Err() != nil {
				return result, fmt.Errorf("send ingest event: %w", ctx.Err())
			}
			if err := waitBeforeRetry(ctx, attempt); err != nil {
				return result, fmt.Errorf("send ingest event: %w", err)
			}
			continue
		}

		if resp.StatusCode() >= 500 {
			if err := waitBeforeRetry(ctx, attempt); err != nil {
				return result, fmt.Errorf("send ingest event: %w", err)
			}
			continue
		}

		if resp.IsError() {
			return result, fmt.Errorf(
				"send ingest event: %s: %s",
				resp.Status(), string(resp.Body()),
			)
		}

		return result, nil
	}
}

func (c *Client) GetAgentConfig(ctx context.Context) (cfg shared.AgentConfig, err error) {
	for attempt := 0; ; attempt++ {
		resp, err := c.resty.R().
			SetContext(ctx).
			SetResult(&cfg).
			Get("/agent/config")
		if err != nil {
			if ctx.Err() != nil {
				return cfg, fmt.Errorf("get agent config: %w", ctx.Err())
			}
			if err := waitBeforeRetry(ctx, attempt); err != nil {
				return cfg, fmt.Errorf("get agent config: %w", err)
			}
			continue
		}

		if resp.StatusCode() >= 500 {
			if err := waitBeforeRetry(ctx, attempt); err != nil {
				return cfg, fmt.Errorf("get agent config: %w", err)
			}
			continue
		}

		if resp.IsError() {
			return cfg, fmt.Errorf(
				"get agent config: %s: %s",
				resp.Status(), string(resp.Body()),
			)
		}

		return cfg, nil
	}
}

func waitBeforeRetry(ctx context.Context, attempt int) error {
	var delay time.Duration
	switch attempt {
	case 0:
		delay = 2 * time.Second
	case 1:
		delay = 4 * time.Second
	default:
		delay = 8 * time.Second
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
