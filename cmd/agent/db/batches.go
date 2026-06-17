package db

import (
	"app/shared"
	"context"
	"encoding/json"
	"time"
)

const sampleRetention = 30 * 24 * time.Hour

func (c *Conn) SaveBatch(ctx context.Context, batch shared.EventBatch) error {
	tx, err := c.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := c.Queries.WithTx(tx)
	for _, event := range batch.Events {
		observedAt := event.ObservedAt.Format(time.RFC3339Nano)

		if event.CPU != nil {
			perCorePercent, err := json.Marshal(event.CPU.PerCorePercent)
			if err != nil {
				return err
			}
			if err := queries.CreateCPUSampleQuery(ctx, CreateCPUSampleQueryParams{
				ObservedAt:     observedAt,
				UsedPercent:    event.CPU.UsedPercent,
				CoresLogical:   int64(event.CPU.CoresLogical),
				CoresPhysical:  int64(event.CPU.CoresPhysical),
				PerCorePercent: string(perCorePercent),
				Load1m:         event.CPU.Load1M,
				Load5m:         event.CPU.Load5M,
				Load15m:        event.CPU.Load15M,
			}); err != nil {
				return err
			}
		}

		if event.Memory != nil {
			if err := queries.CreateMemorySampleQuery(ctx, CreateMemorySampleQueryParams{
				ObservedAt:     observedAt,
				UsedPercent:    event.Memory.UsedPercent,
				UsedBytes:      int64(event.Memory.UsedBytes),
				AvailableBytes: int64(event.Memory.AvailableBytes),
				TotalBytes:     int64(event.Memory.TotalBytes),
			}); err != nil {
				return err
			}
		}

		if event.Disk != nil {
			if err := queries.CreateDiskSampleQuery(ctx, CreateDiskSampleQueryParams{
				ObservedAt:  observedAt,
				Mount:       event.Disk.Mount,
				Filesystem:  event.Disk.Filesystem,
				UsedPercent: event.Disk.UsedPercent,
				UsedBytes:   int64(event.Disk.UsedBytes),
				FreeBytes:   int64(event.Disk.FreeBytes),
				TotalBytes:  int64(event.Disk.TotalBytes),
			}); err != nil {
				return err
			}
		}
	}

	cutoff := batch.SentAt.Add(-sampleRetention).Format(time.RFC3339Nano)
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
