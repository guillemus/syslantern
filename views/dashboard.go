package views

import (
	"fmt"
	"sort"
	"time"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type DashboardStatsData struct {
	HasMetrics           bool
	HostName             string
	HostOS               string
	HostArch             string
	SentAt               time.Time
	CPUUsedPercent       float64
	CPUCoresLogical      int
	MemoryUsedBytes      uint64
	MemoryAvailableBytes uint64
	MemoryTotalBytes     uint64
	Disks                []DashboardDiskData
}

type DashboardDiskData struct {
	Mount       string
	FreeBytes   uint64
	UsedPercent float64
	TotalBytes  uint64
}

func (r *Renderer) Dashboard() Node {
	return Div(
		Class("min-h-dvh bg-zinc-950 p-6 font-mono text-zinc-100"),
		r.DataGet("init", "/dashboard/events"),
		Main(
			Class("mx-auto max-w-5xl space-y-6"),
			DashboardStats(DashboardStatsData{}),
		),
	)
}

func DashboardStats(data DashboardStatsData) Node {
	cpuHeadroom := "—"
	cpuDescription := "Waiting for CPU metrics"
	memoryAvailable := "—"
	memoryDescription := "Waiting for memory metrics"
	if data.HasMetrics {
		cpuHeadroom = percent(100 - data.CPUUsedPercent)
		cpuDescription = fmt.Sprintf("%s currently used across %d cores", percent(data.CPUUsedPercent), data.CPUCoresLogical)
		memoryAvailable = formatBytes(data.MemoryAvailableBytes)
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
		Class("space-y-6"),
		Header(
			Class("flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between"),
			Div(
				P(Class("text-sm text-zinc-500"), Text(headerMeta(data))),
				H1(Class("text-3xl font-semibold"), Text("Resources available")),
			),
			Span(Class("rounded-full bg-emerald-500/15 px-3 py-1 text-sm text-emerald-300"), Text(updatedAt(data.SentAt))),
		),
		Div(
			Class("grid overflow-hidden rounded-xl border border-zinc-800 bg-zinc-900 shadow md:grid-cols-3"),
			availableStat("CPU headroom", cpuHeadroom, cpuDescription, "text-emerald-300"),
			availableStat("RAM available", memoryAvailable, memoryDescription, "text-amber-300"),
			availableStat("Lowest disk free", valueOr(lowestDiskFree, "—"), valueOr(lowestDiskDesc, "Waiting for disk metrics"), "text-red-300"),
		),
		summaryCards(memoryAvailable, memoryDescription, cpuHeadroom, cpuDescription),
		diskTable(disks),
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

func summaryCards(memoryAvailable string, memoryDescription string, cpuHeadroom string, cpuDescription string) Node {
	return Div(
		Class("grid gap-4 md:grid-cols-2"),
		Div(
			Class("rounded-xl border border-zinc-800 bg-zinc-900 p-5"),
			H2(Class("text-xl font-semibold"), Text("Memory")),
			P(Class("mt-3"), Span(Class("text-5xl font-semibold"), Text(memoryAvailable)), Span(Class("ml-2 text-zinc-500"), Text("available"))),
			P(Class("mt-3 text-sm text-zinc-500"), Text(memoryDescription)),
		),
		Div(
			Class("rounded-xl border border-zinc-800 bg-zinc-900 p-5"),
			H2(Class("text-xl font-semibold"), Text("CPU")),
			P(Class("mt-3"), Span(Class("text-5xl font-semibold"), Text(cpuHeadroom)), Span(Class("ml-2 text-zinc-500"), Text("headroom"))),
			P(Class("mt-3 text-sm text-zinc-500"), Text(cpuDescription)),
		),
	)
}

func diskTable(disks []DashboardDiskData) Node {
	if len(disks) == 0 {
		return Div(
			Class("rounded-xl border border-zinc-800 bg-zinc-900 p-5"),
			H2(Class("text-xl font-semibold"), Text("Disks — free space first")),
			P(Class("mt-3 text-zinc-500"), Text("Waiting for disk metrics")),
		)
	}

	rows := make([]Node, 0, len(disks))
	for _, disk := range disks {
		used := disk.UsedPercent
		rows = append(rows, Tr(
			Td(Text(valueOr(disk.Mount, "—"))),
			Td(Class("font-semibold "+diskStatusClass(used)), Text(formatBytes(disk.FreeBytes))),
			Td(Text(percent(used))),
			Td(Text(formatBytes(disk.TotalBytes))),
			Td(Text(diskMeaning(used))),
		))
	}

	return Div(
		Class("rounded-xl border border-zinc-800 bg-zinc-900 p-5"),
		H2(Class("text-xl font-semibold"), Text("Disks — free space first")),
		P(Class("mt-1 text-sm text-zinc-500"), Text("Disk % is used space. Higher means less room left. 100% means full.")),
		Div(
			Class("mt-4 overflow-x-auto"),
			Table(
				Class("w-full text-left text-sm"),
				THead(Tr(
					Th(Class("border-b border-zinc-800 py-2 pr-4 text-zinc-500"), Text("Mount")),
					Th(Class("border-b border-zinc-800 py-2 pr-4 text-zinc-500"), Text("Free")),
					Th(Class("border-b border-zinc-800 py-2 pr-4 text-zinc-500"), Text("Used")),
					Th(Class("border-b border-zinc-800 py-2 pr-4 text-zinc-500"), Text("Total")),
					Th(Class("border-b border-zinc-800 py-2 pr-4 text-zinc-500"), Text("Meaning")),
				)),
				TBody(rows...),
			),
		),
	)
}

func headerMeta(data DashboardStatsData) string {
	if data.HostName == "" {
		return "No host connected"
	}
	hostDetail := fmt.Sprintf("%s %s", data.HostOS, data.HostArch)
	if hostDetail == " " {
		return data.HostName
	}
	return data.HostName + " · " + hostDetail
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

func diskMeaning(used float64) string {
	if used >= 90 {
		return "Near full"
	}
	if used >= 80 {
		return "High usage, not full yet"
	}
	return "Healthy"
}

func diskStatusClass(used float64) string {
	if used >= 90 {
		return "text-red-300"
	}
	if used >= 80 {
		return "text-amber-300"
	}
	return "text-emerald-300"
}

type DashboardExampleResultData struct {
	Count     int
	UpdatedAt string
}

func DashboardExampleResult(data DashboardExampleResultData) Node {
	return Div(
		ID("dashboard-example-result"),
		Class("rounded-lg border border-zinc-800 bg-zinc-950 p-4 text-zinc-300"),
		P(
			Text("Server count is now "),
			Strong(Class("text-zinc-100"), Text(fmt.Sprintf("%d", data.Count))),
			Text("."),
		),
		P(Class("text-sm text-zinc-500"), Text(fmt.Sprintf("Last updated at %s.", data.UpdatedAt))),
	)
}
