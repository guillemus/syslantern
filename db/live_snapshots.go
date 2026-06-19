package db

import (
	"context"
	"encoding/json"
	"syslantern/shared"
	"time"
)

const sampleRetention = 30 * 24 * time.Hour

func (c *Conn) SaveLiveSnapshot(ctx context.Context, snapshot shared.LiveSnapshot) error {
	tx, err := c.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := c.Queries.WithTx(tx)
	metrics := snapshot.Metrics
	observedAt := metrics.ObservedAt.Format(time.RFC3339Nano)

	perCorePercent, err := json.Marshal(metrics.CPU.PerCorePercent)
	if err != nil {
		return err
	}
	if err := queries.CreateCPUSampleQuery(ctx, CreateCPUSampleQueryParams{
		ObservedAt:     observedAt,
		UsedPercent:    metrics.CPU.UsedPercent,
		CoresLogical:   int64(metrics.CPU.CoresLogical),
		CoresPhysical:  int64(metrics.CPU.CoresPhysical),
		PerCorePercent: string(perCorePercent),
		Load1m:         metrics.CPU.Load1M,
		Load5m:         metrics.CPU.Load5M,
		Load15m:        metrics.CPU.Load15M,
	}); err != nil {
		return err
	}

	if err := queries.CreateMemorySampleQuery(ctx, CreateMemorySampleQueryParams{
		ObservedAt:            observedAt,
		VirtualUsedPercent:    metrics.VirtualMemory.UsedPercent,
		VirtualUsedBytes:      int64(metrics.VirtualMemory.UsedBytes),
		VirtualAvailableBytes: int64(metrics.VirtualMemory.AvailableBytes),
		VirtualTotalBytes:     int64(metrics.VirtualMemory.TotalBytes),
		SwapUsedPercent:       metrics.SwapMemory.UsedPercent,
		SwapUsedBytes:         int64(metrics.SwapMemory.UsedBytes),
		SwapAvailableBytes:    int64(metrics.SwapMemory.AvailableBytes),
		SwapTotalBytes:        int64(metrics.SwapMemory.TotalBytes),
	}); err != nil {
		return err
	}

	if err := createDiskSample(ctx, queries, observedAt, true, metrics.Disk.Total); err != nil {
		return err
	}
	for _, partition := range metrics.Disk.Partitions {
		if err := createDiskSample(ctx, queries, observedAt, false, partition); err != nil {
			return err
		}
	}

	cutoff := snapshot.SentAt.Add(-sampleRetention).Format(time.RFC3339Nano)
	if err := queries.DeleteOldCPUSamplesQuery(ctx, cutoff); err != nil {
		return err
	}
	if err := queries.DeleteOldMemorySamplesQuery(ctx, cutoff); err != nil {
		return err
	}
	if err := queries.DeleteOldDiskSamplesQuery(ctx, cutoff); err != nil {
		return err
	}

	return tx.Commit()
}

func createDiskSample(ctx context.Context, queries *Queries, observedAt string, isTotal bool, disk shared.DiskUsage) error {
	var isTotalValue int64
	if isTotal {
		isTotalValue = 1
	}
	return queries.CreateDiskSampleQuery(ctx, CreateDiskSampleQueryParams{
		ObservedAt:  observedAt,
		IsTotal:     isTotalValue,
		Device:      disk.Device,
		Mount:       disk.Mount,
		Filesystem:  disk.Filesystem,
		UsedPercent: disk.UsedPercent,
		UsedBytes:   int64(disk.UsedBytes),
		FreeBytes:   int64(disk.FreeBytes),
		TotalBytes:  int64(disk.TotalBytes),
	})
}
