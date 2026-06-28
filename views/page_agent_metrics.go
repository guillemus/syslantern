package views

import (
	"fmt"
	"sort"
	"time"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func DashboardStats(data DashboardStatsData) Node {
	cpuUsed := "—"
	cpuDescription := "Waiting for CPU metrics"
	memoryUsed := "—"
	memoryDescription := "Waiting for memory metrics"
	if data.HasMetrics {
		cpuUsed = percent(data.CPUUsedPercent)
		cpuDescription = fmt.Sprintf("%d logical cores", data.CPUCoresLogical)
		memoryUsed = percent(memoryUsedPercent(data))
		memoryDescription = fmt.Sprintf("%s used of %s", formatBytes(data.MemoryUsedBytes), formatBytes(data.MemoryTotalBytes))
	}
	disks := sortedDisks(data.Disks)
	lowestDiskFree := ""
	lowestDiskDesc := ""
	if len(disks) > 0 {
		lowest := disks[0]
		lowestDiskFree = percent(diskFreePercent(lowest))
		lowestDiskDesc = fmt.Sprintf("%s has %s free", lowest.Mount, formatBytes(lowest.FreeBytes))
	}

	return Section(
		ID("dashboard-stats"),
		Class("overflow-hidden rounded-lg border border-zinc-700 bg-zinc-950"),
		Div(
			Class("grid md:grid-cols-3"),
			availableStat("CPU used", cpuUsed, cpuDescription, "text-emerald-400"),
			availableStat("Memory used", memoryUsed, memoryDescription, "text-amber-400"),
			availableStat("Lowest disk free", valueOr(lowestDiskFree, "—"), valueOr(lowestDiskDesc, "Waiting for disk metrics"), lowestDiskClass(disks)),
		),
	)
}

func availableStat(title string, value string, desc string, valueClass string) Node {
	return Div(
		Class("border-b border-zinc-800 p-5 md:border-b-0 md:border-r last:border-r-0"),
		P(Class("text-sm text-zinc-500"), Text(title)),
		P(Class("mt-1 text-4xl font-semibold "+valueClass), Text(value)),
		P(Class("mt-1 text-sm text-zinc-500"), Text(desc)),
	)
}

type DashboardDiskData struct {
	Mount       string
	FreeBytes   uint64
	UsedPercent float64
	TotalBytes  uint64
}

type DashboardStatsData struct {
	HasMetrics           bool
	CPUUsedPercent       float64
	CPUCoresLogical      int
	MemoryUsedBytes      uint64
	MemoryAvailableBytes uint64
	MemoryTotalBytes     uint64
	Disks                []DashboardDiskData
}

type DashboardAnalyticsData struct {
	HasAnalytics bool
	Since        time.Time
	CPU          []DashboardCPUHistoryData
	Memory       []DashboardMemoryHistoryData
	Disks        []DashboardDiskHistoryData
}

type DashboardCPUHistoryData struct {
	ObservedAt     time.Time
	UsedPercent    float64
	CoresLogical   int
	CoresPhysical  int
	PerCorePercent []float64
	Load1M         float64
	Load5M         float64
	Load15M        float64
}

type DashboardMemoryHistoryData struct {
	ObservedAt            time.Time
	VirtualUsedPercent    float64
	VirtualUsedBytes      uint64
	VirtualAvailableBytes uint64
	VirtualTotalBytes     uint64
	SwapUsedPercent       float64
	SwapUsedBytes         uint64
	SwapAvailableBytes    uint64
	SwapTotalBytes        uint64
}

type DashboardDiskHistoryData struct {
	ObservedAt  time.Time
	IsTotal     bool
	Mount       string
	Device      string
	Filesystem  string
	UsedPercent float64
	UsedBytes   uint64
	FreeBytes   uint64
	TotalBytes  uint64
}

func valueOr(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func updatedAt(sentAt time.Time) string {
	if sentAt.IsZero() {
		return "waiting"
	}
	return fmt.Sprintf("updated %s", sentAt.Format(time.RFC3339))
}

func sortedDisks(disks []DashboardDiskData) []DashboardDiskData {
	sorted := append([]DashboardDiskData{}, disks...)
	sort.Slice(sorted, func(i, j int) bool {
		return diskFreePercent(sorted[i]) < diskFreePercent(sorted[j])
	})
	return sorted
}

func diskFreePercent(disk DashboardDiskData) float64 {
	return 100 - disk.UsedPercent
}

func memoryUsedPercent(data DashboardStatsData) float64 {
	if data.MemoryTotalBytes == 0 {
		return 0
	}
	return float64(data.MemoryUsedBytes) / float64(data.MemoryTotalBytes) * 100
}

func lowestDiskClass(disks []DashboardDiskData) string {
	if len(disks) == 0 {
		return "text-zinc-400"
	}
	free := diskFreePercent(disks[0])
	if free <= 15 {
		return "text-red-400"
	}
	if free <= 30 {
		return "text-amber-400"
	}
	return "text-emerald-400"
}

func percent(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	value := float64(bytes)
	for _, suffix := range []string{"KB", "MB", "GB", "TB", "PB"} {
		value = value / unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return fmt.Sprintf("%.1f EB", value/unit)
}
