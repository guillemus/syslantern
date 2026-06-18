package db

import (
	"app/shared"
	"context"
	"encoding/json"
	"time"
)

func (c *Conn) LoadAnalytics(ctx context.Context, since time.Time) (shared.AnalyticsSnapshot, error) {
	sinceText := since.Format(time.RFC3339Nano)

	cpuRows, err := c.ListCPUSamplesSinceQuery(ctx, sinceText)
	if err != nil {
		return shared.AnalyticsSnapshot{}, err
	}
	memoryRows, err := c.ListMemorySamplesSinceQuery(ctx, sinceText)
	if err != nil {
		return shared.AnalyticsSnapshot{}, err
	}
	diskRows, err := c.ListDiskSamplesSinceQuery(ctx, sinceText)
	if err != nil {
		return shared.AnalyticsSnapshot{}, err
	}

	snapshot := shared.AnalyticsSnapshot{
		Since:  since,
		CPU:    make([]shared.CPUAnalyticsSample, 0, len(cpuRows)),
		Memory: make([]shared.MemoryAnalyticsSample, 0, len(memoryRows)),
		Disks:  make([]shared.DiskAnalyticsSample, 0, len(diskRows)),
	}

	for _, row := range cpuRows {
		sample, err := cpuAnalyticsSample(row.CpuSample)
		if err != nil {
			return shared.AnalyticsSnapshot{}, err
		}
		snapshot.CPU = append(snapshot.CPU, sample)
	}
	for _, row := range memoryRows {
		sample, err := memoryAnalyticsSample(row.MemorySample)
		if err != nil {
			return shared.AnalyticsSnapshot{}, err
		}
		snapshot.Memory = append(snapshot.Memory, sample)
	}
	for _, row := range diskRows {
		sample, err := diskAnalyticsSample(row.DiskSample)
		if err != nil {
			return shared.AnalyticsSnapshot{}, err
		}
		snapshot.Disks = append(snapshot.Disks, sample)
	}

	return snapshot, nil
}

func cpuAnalyticsSample(row CpuSample) (shared.CPUAnalyticsSample, error) {
	observedAt, err := time.Parse(time.RFC3339Nano, row.ObservedAt)
	if err != nil {
		return shared.CPUAnalyticsSample{}, err
	}
	var perCorePercent []float64
	if err := json.Unmarshal([]byte(row.PerCorePercent), &perCorePercent); err != nil {
		return shared.CPUAnalyticsSample{}, err
	}
	return shared.CPUAnalyticsSample{
		ObservedAt: observedAt,
		CPU: shared.CPUUsage{
			UsedPercent:    row.UsedPercent,
			CoresLogical:   int(row.CoresLogical),
			CoresPhysical:  int(row.CoresPhysical),
			PerCorePercent: perCorePercent,
			Load1M:         row.Load1m,
			Load5M:         row.Load5m,
			Load15M:        row.Load15m,
		},
	}, nil
}

func memoryAnalyticsSample(row MemorySample) (shared.MemoryAnalyticsSample, error) {
	observedAt, err := time.Parse(time.RFC3339Nano, row.ObservedAt)
	if err != nil {
		return shared.MemoryAnalyticsSample{}, err
	}
	return shared.MemoryAnalyticsSample{
		ObservedAt: observedAt,
		VirtualMemory: shared.MemoryUsage{
			UsedPercent:    row.VirtualUsedPercent,
			UsedBytes:      uint64(row.VirtualUsedBytes),
			AvailableBytes: uint64(row.VirtualAvailableBytes),
			TotalBytes:     uint64(row.VirtualTotalBytes),
		},
		SwapMemory: shared.MemoryUsage{
			UsedPercent:    row.SwapUsedPercent,
			UsedBytes:      uint64(row.SwapUsedBytes),
			AvailableBytes: uint64(row.SwapAvailableBytes),
			TotalBytes:     uint64(row.SwapTotalBytes),
		},
	}, nil
}

func diskAnalyticsSample(row DiskSample) (shared.DiskAnalyticsSample, error) {
	observedAt, err := time.Parse(time.RFC3339Nano, row.ObservedAt)
	if err != nil {
		return shared.DiskAnalyticsSample{}, err
	}
	return shared.DiskAnalyticsSample{
		ObservedAt: observedAt,
		IsTotal:    row.IsTotal == 1,
		Disk: shared.DiskUsage{
			Device:      row.Device,
			Mount:       row.Mount,
			Filesystem:  row.Filesystem,
			UsedPercent: row.UsedPercent,
			UsedBytes:   uint64(row.UsedBytes),
			FreeBytes:   uint64(row.FreeBytes),
			TotalBytes:  uint64(row.TotalBytes),
		},
	}, nil
}
