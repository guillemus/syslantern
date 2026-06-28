package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"syslantern/shared"
	"time"
)

const sampleRetention = 30 * 24 * time.Hour

func (c *Conn) SaveLiveSnapshot(ctx context.Context, agentID string, teamID int64, snapshot *shared.LiveSnapshot) (AgentStatus, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("begin live snapshot transaction: %w", err)
	}
	defer tx.Rollback()

	queries := c.WithTx(tx)
	updated, err := queries.setAgentVersion(ctx, setAgentVersionParams{
		ID:      agentID,
		TeamID:  teamID,
		Version: snapshot.Agent.Version,
	})
	if err != nil {
		return "", fmt.Errorf("set agent %s version for team %d: %w", agentID, teamID, err)
	}
	if updated == 0 {
		return "", fmt.Errorf("set agent %s version for team %d: %w", agentID, teamID, sql.ErrNoRows)
	}

	err = queries.setAgentStatus(ctx, setAgentStatusParams{
		Status: AgentStatusRunning,
		ID:     agentID,
		TeamID: teamID,
	})
	if err != nil {
		return "", fmt.Errorf("set agent %s running status for team %d: %w", agentID, teamID, err)
	}

	metrics := snapshot.Metrics
	observedAt := metrics.ObservedAt.Format(time.RFC3339Nano)
	virtualUsedBytes, err := int64FromUint64(metrics.VirtualMemory.UsedBytes)
	if err != nil {
		return "", fmt.Errorf("convert virtual used bytes: %w", err)
	}
	virtualAvailableBytes, err := int64FromUint64(metrics.VirtualMemory.AvailableBytes)
	if err != nil {
		return "", fmt.Errorf("convert virtual available bytes: %w", err)
	}
	virtualTotalBytes, err := int64FromUint64(metrics.VirtualMemory.TotalBytes)
	if err != nil {
		return "", fmt.Errorf("convert virtual total bytes: %w", err)
	}
	swapUsedBytes, err := int64FromUint64(metrics.SwapMemory.UsedBytes)
	if err != nil {
		return "", fmt.Errorf("convert swap used bytes: %w", err)
	}
	swapAvailableBytes, err := int64FromUint64(metrics.SwapMemory.AvailableBytes)
	if err != nil {
		return "", fmt.Errorf("convert swap available bytes: %w", err)
	}
	swapTotalBytes, err := int64FromUint64(metrics.SwapMemory.TotalBytes)
	if err != nil {
		return "", fmt.Errorf("convert swap total bytes: %w", err)
	}

	perCorePercent, err := json.Marshal(metrics.CPU.PerCorePercent)
	if err != nil {
		return "", fmt.Errorf("marshal CPU per-core percent: %w", err)
	}
	if err := queries.createCPUSample(ctx, createCPUSampleParams{
		ObservedAt:     observedAt,
		UsedPercent:    metrics.CPU.UsedPercent,
		CoresLogical:   int64(metrics.CPU.CoresLogical),
		CoresPhysical:  int64(metrics.CPU.CoresPhysical),
		PerCorePercent: string(perCorePercent),
		Load1m:         metrics.CPU.Load1M,
		Load5m:         metrics.CPU.Load5M,
		Load15m:        metrics.CPU.Load15M,
	}); err != nil {
		return "", fmt.Errorf("create CPU sample: %w", err)
	}

	if err := queries.createMemorySample(ctx, createMemorySampleParams{
		ObservedAt:            observedAt,
		VirtualUsedPercent:    metrics.VirtualMemory.UsedPercent,
		VirtualUsedBytes:      virtualUsedBytes,
		VirtualAvailableBytes: virtualAvailableBytes,
		VirtualTotalBytes:     virtualTotalBytes,
		SwapUsedPercent:       metrics.SwapMemory.UsedPercent,
		SwapUsedBytes:         swapUsedBytes,
		SwapAvailableBytes:    swapAvailableBytes,
		SwapTotalBytes:        swapTotalBytes,
	}); err != nil {
		return "", fmt.Errorf("create memory sample: %w", err)
	}

	if err := saveDiskSample(ctx, queries, observedAt, true, metrics.Disk.Total); err != nil {
		return "", fmt.Errorf("create total disk sample: %w", err)
	}
	for _, partition := range metrics.Disk.Partitions {
		if err := saveDiskSample(ctx, queries, observedAt, false, partition); err != nil {
			return "", fmt.Errorf("create disk sample for mount %s: %w", partition.Mount, err)
		}
	}

	cutoff := snapshot.SentAt.Add(-sampleRetention).Format(time.RFC3339Nano)
	if err := queries.deleteOldCPUSamples(ctx, cutoff); err != nil {
		return "", fmt.Errorf("delete old CPU samples: %w", err)
	}
	if err := queries.deleteOldMemorySamples(ctx, cutoff); err != nil {
		return "", fmt.Errorf("delete old memory samples: %w", err)
	}
	if err := queries.deleteOldDiskSamples(ctx, cutoff); err != nil {
		return "", fmt.Errorf("delete old disk samples: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit live snapshot transaction: %w", err)
	}

	return AgentStatusRunning, nil
}

func saveDiskSample(ctx context.Context, queries *Queries, observedAt string, isTotal bool, disk shared.DiskUsage) error {
	var isTotalValue int64
	if isTotal {
		isTotalValue = 1
	}
	usedBytes, err := int64FromUint64(disk.UsedBytes)
	if err != nil {
		return fmt.Errorf("convert disk used bytes: %w", err)
	}
	freeBytes, err := int64FromUint64(disk.FreeBytes)
	if err != nil {
		return fmt.Errorf("convert disk free bytes: %w", err)
	}
	totalBytes, err := int64FromUint64(disk.TotalBytes)
	if err != nil {
		return fmt.Errorf("convert disk total bytes: %w", err)
	}
	return queries.createDiskSample(ctx, createDiskSampleParams{
		ObservedAt:  observedAt,
		IsTotal:     isTotalValue,
		Device:      disk.Device,
		Mount:       disk.Mount,
		Filesystem:  disk.Filesystem,
		UsedPercent: disk.UsedPercent,
		UsedBytes:   usedBytes,
		FreeBytes:   freeBytes,
		TotalBytes:  totalBytes,
	})
}

func int64FromUint64(value uint64) (int64, error) {
	if value > math.MaxInt64 {
		return 0, fmt.Errorf("%d exceeds int64 max", value)
	}
	return int64(value), nil
}
