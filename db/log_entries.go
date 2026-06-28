package db

import (
	"context"
	"fmt"
	"syslantern/shared"
	"time"
)

func (c *Conn) SaveLogs(ctx context.Context, agentID string, teamID int64, logs []shared.LogEvent) (AgentStatus, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin logs transaction: %w", err)
	}
	defer tx.Rollback()

	queries := c.WithTx(tx)
	if err := saveLogEntries(ctx, queries, agentID, teamID, logs); err != nil {
		return "", err
	}

	if err := queries.setAgentStatus(ctx, setAgentStatusParams{
		Status: AgentStatusRunning,
		ID:     agentID,
		TeamID: teamID,
	}); err != nil {
		return "", fmt.Errorf("set agent %s running status for team %d: %w", agentID, teamID, err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit logs transaction: %w", err)
	}

	return AgentStatusRunning, nil
}

func saveLogEntries(ctx context.Context, queries *Queries, agentID string, teamID int64, logs []shared.LogEvent) error {
	receivedAt := time.Now().UTC().Format(time.RFC3339Nano)
	for _, log := range logs {
		if err := queries.createLogEntry(ctx, createLogEntryParams{
			ID:         log.ID,
			TeamID:     teamID,
			AgentID:    agentID,
			ObservedAt: log.ObservedAt.Format(time.RFC3339Nano),
			ReceivedAt: receivedAt,
			Source:     log.Source,
			Unit:       log.Unit,
			Priority:   log.Priority,
			Message:    log.Message,
		}); err != nil {
			return fmt.Errorf("create log entry %s: %w", log.ID, err)
		}
	}

	return nil
}
