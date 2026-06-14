package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func NewWidgetID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("widget id: %w", err)
	}
	return "wid_" + hex.EncodeToString(b[:]), nil
}

func (c *Conn) GetAgent(ctx context.Context, workspaceID, agentID int64) (Agent, error) {
	row, err := c.GetAgentQuery(ctx, GetAgentQueryParams{
		WorkspaceID: workspaceID,
		AgentID:     agentID,
	})
	return row.Agent, err
}

func (c *Conn) CreateAgent(ctx context.Context, workspaceID int64, name string) (Agent, error) {
	widgetID, err := NewWidgetID()
	if err != nil {
		return Agent{}, err
	}
	return c.CreateAgentQuery(ctx, CreateAgentQueryParams{WorkspaceID: workspaceID, WidgetID: widgetID, Name: name})
}

func (c *Conn) WidgetExists(ctx context.Context, widgetID string) error {
	_, err := c.WidgetExistsQuery(ctx, widgetID)
	return err
}
