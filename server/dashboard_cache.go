package server

import (
	"app/shared"
	"app/views"
	"sort"
	"time"
)

type DashboardCache struct {
	items *ConcurrentMap[shared.AgentID, views.DashboardData]
}

func NewDashboardCache() *DashboardCache {
	return &DashboardCache{
		items: NewConcurrentMap[shared.AgentID, views.DashboardData](),
	}
}

func (c *DashboardCache) Get(agentID shared.AgentID) (views.DashboardData, bool) {
	return c.items.Get(agentID)
}

func (c *DashboardCache) List() []views.DashboardData {
	items := []views.DashboardData{}
	c.items.Range(func(_ shared.AgentID, value views.DashboardData) bool {
		items = append(items, value)
		return true
	})
	sort.Slice(items, func(i, j int) bool {
		return items[i].AgentID < items[j].AgentID
	})
	return items
}

func (c *DashboardCache) UpsertLiveSnapshot(snapshot shared.LiveSnapshot) views.DashboardData {
	data, _ := c.items.Get(snapshot.Agent.ID)
	data.AgentID = string(snapshot.Agent.ID)
	data.Stats = dashboardStatsFromLiveSnapshot(snapshot)
	data.Analytics = appendLiveSnapshotToAnalytics(data.Analytics, snapshot)
	c.items.Set(snapshot.Agent.ID, data)
	return data
}

func (c *DashboardCache) UpsertAnalytics(snapshot shared.AnalyticsSnapshot) views.DashboardData {
	data, _ := c.items.Get(snapshot.Agent.ID)
	data.AgentID = string(snapshot.Agent.ID)
	data.Analytics = dashboardAnalyticsFromSnapshot(snapshot)
	c.items.Set(snapshot.Agent.ID, data)
	return data
}

func appendLiveSnapshotToAnalytics(data views.DashboardAnalyticsData, snapshot shared.LiveSnapshot) views.DashboardAnalyticsData {
	if !data.HasAnalytics {
		data.HasAnalytics = true
		data.Since = snapshot.SentAt.Add(-1 * time.Hour)
	}
	data.SentAt = snapshot.SentAt

	metrics := snapshot.Metrics
	data.CPU = append(data.CPU, views.DashboardCPUHistoryData{
		ObservedAt:     metrics.ObservedAt,
		UsedPercent:    metrics.CPU.UsedPercent,
		CoresLogical:   metrics.CPU.CoresLogical,
		CoresPhysical:  metrics.CPU.CoresPhysical,
		PerCorePercent: metrics.CPU.PerCorePercent,
		Load1M:         metrics.CPU.Load1M,
		Load5M:         metrics.CPU.Load5M,
		Load15M:        metrics.CPU.Load15M,
	})
	data.Memory = append(data.Memory, views.DashboardMemoryHistoryData{
		ObservedAt:            metrics.ObservedAt,
		VirtualUsedPercent:    metrics.VirtualMemory.UsedPercent,
		VirtualUsedBytes:      metrics.VirtualMemory.UsedBytes,
		VirtualAvailableBytes: metrics.VirtualMemory.AvailableBytes,
		VirtualTotalBytes:     metrics.VirtualMemory.TotalBytes,
		SwapUsedPercent:       metrics.SwapMemory.UsedPercent,
		SwapUsedBytes:         metrics.SwapMemory.UsedBytes,
		SwapAvailableBytes:    metrics.SwapMemory.AvailableBytes,
		SwapTotalBytes:        metrics.SwapMemory.TotalBytes,
	})
	data.Disks = append(data.Disks, dashboardDiskHistoryFromUsage(metrics.ObservedAt, true, metrics.Disk.Total))
	for _, disk := range metrics.Disk.Partitions {
		data.Disks = append(data.Disks, dashboardDiskHistoryFromUsage(metrics.ObservedAt, false, disk))
	}

	return trimDashboardAnalytics(data)
}

func dashboardDiskHistoryFromUsage(observedAt time.Time, isTotal bool, disk shared.DiskUsage) views.DashboardDiskHistoryData {
	return views.DashboardDiskHistoryData{
		ObservedAt:  observedAt,
		IsTotal:     isTotal,
		Mount:       disk.Mount,
		Device:      disk.Device,
		Filesystem:  disk.Filesystem,
		UsedPercent: disk.UsedPercent,
		UsedBytes:   disk.UsedBytes,
		FreeBytes:   disk.FreeBytes,
		TotalBytes:  disk.TotalBytes,
	}
}

func trimDashboardAnalytics(data views.DashboardAnalyticsData) views.DashboardAnalyticsData {
	data.CPU = trimCPUHistory(data.CPU, data.Since)
	data.Memory = trimMemoryHistory(data.Memory, data.Since)
	data.Disks = trimDiskHistory(data.Disks, data.Since)
	return data
}

func trimCPUHistory(samples []views.DashboardCPUHistoryData, since time.Time) []views.DashboardCPUHistoryData {
	filtered := samples[:0]
	for _, sample := range samples {
		if sample.ObservedAt.Before(since) {
			continue
		}
		filtered = append(filtered, sample)
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ObservedAt.Before(filtered[j].ObservedAt)
	})
	return filtered
}

func trimMemoryHistory(samples []views.DashboardMemoryHistoryData, since time.Time) []views.DashboardMemoryHistoryData {
	filtered := samples[:0]
	for _, sample := range samples {
		if sample.ObservedAt.Before(since) {
			continue
		}
		filtered = append(filtered, sample)
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ObservedAt.Before(filtered[j].ObservedAt)
	})
	return filtered
}

func trimDiskHistory(samples []views.DashboardDiskHistoryData, since time.Time) []views.DashboardDiskHistoryData {
	filtered := samples[:0]
	for _, sample := range samples {
		if sample.ObservedAt.Before(since) {
			continue
		}
		filtered = append(filtered, sample)
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].ObservedAt.Equal(filtered[j].ObservedAt) {
			return filtered[i].Mount < filtered[j].Mount
		}
		return filtered[i].ObservedAt.Before(filtered[j].ObservedAt)
	})
	return filtered
}

func dashboardStatsFromLiveSnapshot(snapshot shared.LiveSnapshot) views.DashboardStatsData {
	metrics := snapshot.Metrics
	disks := append([]shared.DiskUsage{metrics.Disk.Total}, metrics.Disk.Partitions...)
	viewDisks := make([]views.DashboardDiskData, 0, len(disks))

	for _, disk := range disks {
		viewDisks = append(viewDisks, views.DashboardDiskData{
			Mount:       disk.Mount,
			FreeBytes:   disk.FreeBytes,
			UsedPercent: disk.UsedPercent,
			TotalBytes:  disk.TotalBytes,
		})
	}

	return views.DashboardStatsData{
		HasMetrics:           true,
		HostName:             snapshot.Host.Name,
		HostOS:               snapshot.Host.OS,
		HostArch:             snapshot.Host.Arch,
		SentAt:               snapshot.SentAt,
		CPUUsedPercent:       metrics.CPU.UsedPercent,
		CPUCoresLogical:      metrics.CPU.CoresLogical,
		MemoryUsedBytes:      metrics.VirtualMemory.UsedBytes,
		MemoryAvailableBytes: metrics.VirtualMemory.AvailableBytes,
		MemoryTotalBytes:     metrics.VirtualMemory.TotalBytes,
		Disks:                viewDisks,
	}
}

func dashboardAnalyticsFromSnapshot(snapshot shared.AnalyticsSnapshot) views.DashboardAnalyticsData {
	data := views.DashboardAnalyticsData{
		HasAnalytics: true,
		SentAt:       snapshot.SentAt,
		Since:        snapshot.Since,
		CPU:          make([]views.DashboardCPUHistoryData, 0, len(snapshot.CPU)),
		Memory:       make([]views.DashboardMemoryHistoryData, 0, len(snapshot.Memory)),
		Disks:        make([]views.DashboardDiskHistoryData, 0, len(snapshot.Disks)),
	}

	for _, sample := range snapshot.CPU {
		data.CPU = append(data.CPU, views.DashboardCPUHistoryData{
			ObservedAt:     sample.ObservedAt,
			UsedPercent:    sample.CPU.UsedPercent,
			CoresLogical:   sample.CPU.CoresLogical,
			CoresPhysical:  sample.CPU.CoresPhysical,
			PerCorePercent: sample.CPU.PerCorePercent,
			Load1M:         sample.CPU.Load1M,
			Load5M:         sample.CPU.Load5M,
			Load15M:        sample.CPU.Load15M,
		})
	}
	for _, sample := range snapshot.Memory {
		data.Memory = append(data.Memory, views.DashboardMemoryHistoryData{
			ObservedAt:            sample.ObservedAt,
			VirtualUsedPercent:    sample.VirtualMemory.UsedPercent,
			VirtualUsedBytes:      sample.VirtualMemory.UsedBytes,
			VirtualAvailableBytes: sample.VirtualMemory.AvailableBytes,
			VirtualTotalBytes:     sample.VirtualMemory.TotalBytes,
			SwapUsedPercent:       sample.SwapMemory.UsedPercent,
			SwapUsedBytes:         sample.SwapMemory.UsedBytes,
			SwapAvailableBytes:    sample.SwapMemory.AvailableBytes,
			SwapTotalBytes:        sample.SwapMemory.TotalBytes,
		})
	}
	for _, sample := range snapshot.Disks {
		data.Disks = append(data.Disks, views.DashboardDiskHistoryData{
			ObservedAt:  sample.ObservedAt,
			IsTotal:     sample.IsTotal,
			Mount:       sample.Disk.Mount,
			Device:      sample.Disk.Device,
			Filesystem:  sample.Disk.Filesystem,
			UsedPercent: sample.Disk.UsedPercent,
			UsedBytes:   sample.Disk.UsedBytes,
			FreeBytes:   sample.Disk.FreeBytes,
			TotalBytes:  sample.Disk.TotalBytes,
		})
	}

	return data
}
