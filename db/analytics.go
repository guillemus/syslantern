package db

import (
	"encoding/json"
	"syslantern/shared"
	"time"
)

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
