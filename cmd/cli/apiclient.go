package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"app/shared"

	"github.com/bytedance/sonic"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: "http://host.multipass:3000",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SendBatch(ctx context.Context, batch shared.EventBatch) error {
	body, err := sonic.Marshal(batch)
	if err != nil {
		return fmt.Errorf("encode event batch: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/event-batches", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create event batch request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send event batch: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("send event batch: %s", response.Status)
		}
		return fmt.Errorf("send event batch: %s: %s", response.Status, strings.TrimSpace(string(responseBody)))
	}

	return nil
}
