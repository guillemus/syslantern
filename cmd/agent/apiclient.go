package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"syslantern/shared"

	"github.com/bytedance/sonic"
	"github.com/go-resty/resty/v2"
)

type Client struct {
	resty *resty.Client
}

func NewClient() *Client {
	return &Client{
		resty: resty.New().
			SetBaseURL("http://host.multipass:3000"),
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

func (c *Client) Connect(ctx context.Context, agentID shared.AgentID) <-chan shared.Command {
	commands := make(chan shared.Command)

	go func() {
		defer close(commands)

		for {
			if err := ctx.Err(); err != nil {
				return
			}
			err := c.streamCommands(ctx, agentID, commands)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			select {
			case <-ctx.Done():
			case <-time.After(2 * time.Second):
			}
		}
	}()

	return commands
}

func (c *Client) streamCommands(ctx context.Context, agentID shared.AgentID, commands chan<- shared.Command) error {
	resp, err := c.resty.R().
		SetContext(ctx).
		SetDoNotParseResponse(true).
		SetQueryParam("agent_id", string(agentID)).
		Get("/connect")
	if err != nil {
		return fmt.Errorf("open command stream: %w", err)
	}
	defer resp.RawBody().Close()

	if resp.IsError() {
		return fmt.Errorf("open command stream: %s", resp.Status())
	}

	decoder := sonic.ConfigDefault.NewDecoder(resp.RawBody())
	for {
		var command shared.Command
		if err := decoder.Decode(&command); err != nil {
			return fmt.Errorf("read command stream: %w", err)
		}

		select {
		case <-ctx.Done():
			return nil
		case commands <- command:
		}
	}
}
