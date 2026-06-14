package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func NewVisitorID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("visitor id: %w", err)
	}
	return "vis_" + hex.EncodeToString(b[:]), nil
}

func (c *Conn) EnsureVisitorConversation(ctx context.Context, widgetID, visitorID string) (Conversation, bool, error) {
	if err := c.WidgetExists(ctx, widgetID); err != nil {
		return Conversation{}, false, err
	}

	created := false
	if visitorID == "" {
		id, err := NewVisitorID()
		if err != nil {
			return Conversation{}, false, err
		}
		visitorID = id
		created = true
	}

	if err := c.UpsertVisitorQuery(ctx, UpsertVisitorQueryParams{VisitorID: visitorID, WidgetID: widgetID}); err != nil {
		return Conversation{}, false, err
	}

	conversation, err := c.UpsertConversationQuery(ctx, visitorID)
	if err != nil {
		return Conversation{}, false, err
	}

	return conversation, created, nil
}

func (c *Conn) CreateMessage(ctx context.Context, conversationID int64, role, text string) error {
	if err := c.CreateMessageQuery(ctx, CreateMessageQueryParams{ConversationID: conversationID, Role: role, Text: text}); err != nil {
		return err
	}
	return c.TouchConversationQuery(ctx, conversationID)
}

func (c *Conn) ListMessages(ctx context.Context, conversationID int64) ([]Message, error) {
	rows, err := c.ListMessagesQuery(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	messages := make([]Message, len(rows))
	for i, row := range rows {
		messages[i] = row.Message
	}
	return messages, nil
}
