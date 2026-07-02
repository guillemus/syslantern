package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"syslantern/shared"
	"time"
)

type journalLogEntry struct {
	Cursor            string `json:"__CURSOR"`
	RealtimeTimestamp string `json:"__REALTIME_TIMESTAMP"`
	SystemdUnit       string `json:"_SYSTEMD_UNIT"`
	SystemdUserUnit   string `json:"_SYSTEMD_USER_UNIT"`
	Priority          string `json:"PRIORITY"`
	Message           string `json:"MESSAGE"`
}

func (a *Agent) collectJournalLogs(
	ctx context.Context, host shared.Host, cursor string, limit int,
) ([]shared.LogEvent, string, error) {
	args := []string{"-o", "json", "--no-pager", "-n", strconv.Itoa(limit), "--after-cursor", cursor}
	if cursor == "" {
		args = []string{"-o", "json", "--no-pager", "-n", "1"}
	}

	cmd := exec.CommandContext(ctx, "journalctl", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, cursor, fmt.Errorf("journalctl: %w", err)
	}

	now := time.Now().UTC()
	logs := make([]shared.LogEvent, 0)
	lastCursor := cursor
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		var entry journalLogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return nil, cursor, fmt.Errorf("parse journal entry: %w", err)
		}
		if entry.Cursor == "" || entry.Message == "" {
			continue
		}

		observedAt, err := journalTimestamp(entry.RealtimeTimestamp)
		if err != nil {
			a.logger.Warn("failed to parse journal timestamp", "timestamp", entry.RealtimeTimestamp, "error", err)
			continue
		}

		unit := entry.SystemdUnit
		if unit == "" {
			unit = entry.SystemdUserUnit
		}

		if cursor != "" {
			logs = append(logs, shared.LogEvent{
				ID:         "log_" + eventID(observedAt, len(logs)),
				Host:       host,
				SentAt:     now,
				ObservedAt: observedAt,
				Source:     "journal",
				Metadata: map[string]string{
					"unit":     unit,
					"priority": entry.Priority,
				},
				Message: entry.Message,
			})
		}
		lastCursor = entry.Cursor
	}
	if err := scanner.Err(); err != nil {
		return nil, cursor, fmt.Errorf("scan journal output: %w", err)
	}

	return logs, lastCursor, nil
}

func journalTimestamp(value string) (time.Time, error) {
	microseconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMicro(microseconds).UTC(), nil
}
