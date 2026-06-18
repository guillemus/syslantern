package views

import (
	"app/shared"
	"fmt"
	"sort"
	"time"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type DashboardStatsData struct {
	HostName          string
	HostDetail        string
	UpdatedAt         string
	CPUHeadroom       string
	CPUDescription    string
	MemoryAvailable   string
	MemoryDescription string
	LowestDiskFree    string
	LowestDiskDesc    string
	Disks             []DashboardDiskData
}

type DashboardDiskData struct {
	Mount            string
	Free             string
	FreeBytes        uint64
	FreePercent      string
	FreePercentValue float64
	Used             string
	Total            string
	Meaning          string
	StatusClass      string
}

func DashboardStatsFromBatch(batch shared.EventBatch) DashboardStatsData {
	stats := DashboardStatsData{
		HostName:   batch.Host.Name,
		HostDetail: fmt.Sprintf("%s %s", batch.Host.OS, batch.Host.Arch),
		UpdatedAt:  fmt.Sprintf("updated %s", batch.SentAt.Format(time.RFC3339)),
		Disks:      []DashboardDiskData{},
	}

	metrics := batch.Metrics
	stats.CPUHeadroom = percent(100 - metrics.CPU.UsedPercent)
	stats.CPUDescription = fmt.Sprintf("%s currently used across %d cores", percent(metrics.CPU.UsedPercent), metrics.CPU.CoresLogical)
	stats.MemoryAvailable = formatBytes(metrics.VirtualMemory.AvailableBytes)
	stats.MemoryDescription = fmt.Sprintf("%s used of %s", formatBytes(metrics.VirtualMemory.UsedBytes), formatBytes(metrics.VirtualMemory.TotalBytes))

	disks := append([]shared.DiskUsage{metrics.Disk.Total}, metrics.Disk.Partitions...)
	for _, disk := range disks {
		used := disk.UsedPercent
		freePercent := 100 - used
		stats.Disks = append(stats.Disks, DashboardDiskData{
			Mount:            disk.Mount,
			Free:             formatBytes(disk.FreeBytes),
			FreeBytes:        disk.FreeBytes,
			FreePercent:      percent(freePercent),
			FreePercentValue: freePercent,
			Used:             percent(used),
			Total:            formatBytes(disk.TotalBytes),
			Meaning:          diskMeaning(used),
			StatusClass:      diskStatusClass(used),
		})
	}

	sort.Slice(stats.Disks, func(i, j int) bool {
		return stats.Disks[i].FreePercentValue < stats.Disks[j].FreePercentValue
	})

	if len(stats.Disks) > 0 {
		lowest := stats.Disks[0]
		stats.LowestDiskFree = lowest.FreePercent
		stats.LowestDiskDesc = fmt.Sprintf("%s has %s free", lowest.Mount, lowest.Free)
	}

	return stats
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
	return Section(
		ID("dashboard-stats"),
		Class("space-y-6"),
		Header(
			Class("flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between"),
			Div(
				P(Class("text-sm text-zinc-500"), Text(headerMeta(data))),
				H1(Class("text-3xl font-semibold"), Text("Resources available")),
			),
			Span(Class("rounded-full bg-emerald-500/15 px-3 py-1 text-sm text-emerald-300"), Text(valueOr(data.UpdatedAt, "waiting"))),
		),
		Div(
			Class("grid overflow-hidden rounded-xl border border-zinc-800 bg-zinc-900 shadow md:grid-cols-3"),
			availableStat("CPU headroom", valueOr(data.CPUHeadroom, "—"), valueOr(data.CPUDescription, "Waiting for CPU metrics"), "text-emerald-300"),
			availableStat("RAM available", valueOr(data.MemoryAvailable, "—"), valueOr(data.MemoryDescription, "Waiting for memory metrics"), "text-amber-300"),
			availableStat("Lowest disk free", valueOr(data.LowestDiskFree, "—"), valueOr(data.LowestDiskDesc, "Waiting for disk metrics"), "text-red-300"),
		),
		summaryCards(data),
		diskTable(data.Disks),
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

func summaryCards(data DashboardStatsData) Node {
	return Div(
		Class("grid gap-4 md:grid-cols-2"),
		Div(
			Class("rounded-xl border border-zinc-800 bg-zinc-900 p-5"),
			H2(Class("text-xl font-semibold"), Text("Memory")),
			P(Class("mt-3"), Span(Class("text-5xl font-semibold"), Text(valueOr(data.MemoryAvailable, "—"))), Span(Class("ml-2 text-zinc-500"), Text("available"))),
			P(Class("mt-3 text-sm text-zinc-500"), Text(valueOr(data.MemoryDescription, "Waiting for memory metrics"))),
		),
		Div(
			Class("rounded-xl border border-zinc-800 bg-zinc-900 p-5"),
			H2(Class("text-xl font-semibold"), Text("CPU")),
			P(Class("mt-3"), Span(Class("text-5xl font-semibold"), Text(valueOr(data.CPUHeadroom, "—"))), Span(Class("ml-2 text-zinc-500"), Text("headroom"))),
			P(Class("mt-3 text-sm text-zinc-500"), Text(valueOr(data.CPUDescription, "Waiting for CPU metrics"))),
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
		rows = append(rows, Tr(
			Td(Text(valueOr(disk.Mount, "—"))),
			Td(Class("font-semibold "+disk.StatusClass), Text(disk.Free)),
			Td(Text(disk.Used)),
			Td(Text(disk.Total)),
			Td(Text(disk.Meaning)),
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
	if data.HostDetail == "" {
		return data.HostName
	}
	return data.HostName + " · " + data.HostDetail
}

func valueOr(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
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
