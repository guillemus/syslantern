package views

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"time"

	historyscripts "app/views/scripts"

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

type DashboardData struct {
	AgentID   string
	Stats     DashboardStatsData
	Analytics DashboardAnalyticsData
}

type DashboardAnalyticsData struct {
	HasAnalytics bool
	SentAt       time.Time
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

type DashboardDiskData struct {
	Mount       string
	FreeBytes   uint64
	UsedPercent float64
	TotalBytes  uint64
}

func (r *Renderer) AgentsIndex(data []DashboardData) Node {
	rows := make([]Node, 0, len(data))
	for _, dashboard := range data {
		rows = append(rows, r.machineRow(dashboard))
	}
	if len(rows) == 0 {
		rows = append(rows, Tr(
			Td(ColSpan("5"), Class("py-6 text-zinc-500"), Text("No machines connected yet.")),
		))
	}

	return Div(
		Class("min-h-dvh bg-zinc-950 p-6 font-mono text-zinc-100"),
		Main(
			Class("mx-auto max-w-5xl space-y-6"),
			Header(
				H1(Class("text-3xl font-semibold"), Text("Machines")),
				P(Class("mt-2 text-sm text-zinc-500"), Text("Connected agents with cached dashboard state.")),
			),
			Div(
				Class("overflow-hidden rounded-xl border border-zinc-800 bg-zinc-900"),
				Table(
					Class("w-full text-left text-sm"),
					THead(Tr(
						Th(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text("Host")),
						Th(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text("Agent")),
						Th(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text("CPU")),
						Th(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text("Memory")),
						Th(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text("Updated")),
					)),
					TBody(rows...),
				),
			),
		),
	)
}

func (r *Renderer) machineRow(data DashboardData) Node {
	href := r.URL("GET", "/agents/"+url.PathEscape(data.AgentID))
	host := valueOr(data.Stats.HostName, "unknown")
	cpu := "—"
	memory := "—"
	if data.Stats.HasMetrics {
		cpu = percent(data.Stats.CPUUsedPercent)
		memory = fmt.Sprintf("%s / %s", formatBytes(data.Stats.MemoryUsedBytes), formatBytes(data.Stats.MemoryTotalBytes))
	}
	return Tr(
		Td(Class("border-b border-zinc-800 px-4 py-3"), A(Href(href), Class("font-semibold text-zinc-100 hover:text-emerald-300"), Text(host))),
		Td(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text(data.AgentID)),
		Td(Class("border-b border-zinc-800 px-4 py-3"), Text(cpu)),
		Td(Class("border-b border-zinc-800 px-4 py-3"), Text(memory)),
		Td(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text(updatedAt(data.Stats.SentAt))),
	)
}

func (r *Renderer) Dashboard(data DashboardData) Node {
	return Div(
		Class("min-h-dvh bg-zinc-950 p-4 font-mono text-zinc-100 sm:p-6"),
		Data("signals", dashboardHistorySignals(data)),
		r.dashboardEventsDataGet(data),
		Main(
			Class("mx-auto flex max-w-7xl flex-col gap-4"),
			dashboardHeader(data),
			DashboardStats(data.Stats),
			Section(
				Class("grid gap-4 xl:grid-cols-[1.5fr_1fr]"),
				historyscripts.CPUHistory("$dashboardHistory.cpu"),
				historyscripts.MemoryPressure("$dashboardHistory.memory"),
			),
			historyscripts.DiskPressure("$dashboardHistory.disks"),
		),
	)
}

func (r *Renderer) dashboardEventsDataGet(data DashboardData) Node {
	if data.AgentID == "" {
		return Text("")
	}
	return r.DataGet("init", "/agents/"+url.PathEscape(data.AgentID)+"/events")
}

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

func dashboardHeader(data DashboardData) Node {
	host := valueOr(data.Stats.HostName, "waiting for agent")
	return Header(
		Class("flex flex-col gap-3 border-b border-zinc-700 pb-4 lg:flex-row lg:items-end lg:justify-between"),
		Div(
			Div(
				Class("flex flex-wrap items-center gap-2 text-sm text-zinc-500"),
				Span(Class("rounded border border-zinc-700 px-2 py-0.5"), Text(valueOr(data.AgentID, "agent"))),
				Span(Text(headerMeta(data.Stats))),
			),
			H1(Class("mt-2 text-2xl font-semibold tracking-normal sm:text-3xl"), Text(host)),
		),
		Span(Class("text-sm text-zinc-500"), Text(updatedAt(data.Stats.SentAt))),
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

type dashboardHistorySignalsData struct {
	DashboardHistory dashboardHistorySignal `json:"dashboardHistory"`
}

type dashboardHistorySignal struct {
	CPU    []dashboardCPUHistorySignal    `json:"cpu"`
	Memory []dashboardMemoryHistorySignal `json:"memory"`
	Disks  []dashboardDiskHistorySignal   `json:"disks"`
}

type dashboardCPUHistorySignal struct {
	ObservedAt     time.Time `json:"observedAt"`
	UsedPercent    float64   `json:"usedPercent"`
	CoresLogical   int       `json:"coresLogical"`
	CoresPhysical  int       `json:"coresPhysical"`
	PerCorePercent []float64 `json:"perCorePercent"`
	Load1M         float64   `json:"load1M"`
	Load5M         float64   `json:"load5M"`
	Load15M        float64   `json:"load15M"`
}

type dashboardMemoryHistorySignal struct {
	ObservedAt            time.Time `json:"observedAt"`
	VirtualUsedPercent    float64   `json:"virtualUsedPercent"`
	VirtualUsedBytes      uint64    `json:"virtualUsedBytes"`
	VirtualAvailableBytes uint64    `json:"virtualAvailableBytes"`
	VirtualTotalBytes     uint64    `json:"virtualTotalBytes"`
	SwapUsedPercent       float64   `json:"swapUsedPercent"`
	SwapUsedBytes         uint64    `json:"swapUsedBytes"`
	SwapAvailableBytes    uint64    `json:"swapAvailableBytes"`
	SwapTotalBytes        uint64    `json:"swapTotalBytes"`
}

type dashboardDiskHistorySignal struct {
	ObservedAt  time.Time `json:"observedAt"`
	IsTotal     bool      `json:"isTotal"`
	Mount       string    `json:"mount"`
	Device      string    `json:"device"`
	Filesystem  string    `json:"filesystem"`
	UsedPercent float64   `json:"usedPercent"`
	UsedBytes   uint64    `json:"usedBytes"`
	FreeBytes   uint64    `json:"freeBytes"`
	TotalBytes  uint64    `json:"totalBytes"`
}

func dashboardHistorySignals(data DashboardData) string {
	b, _ := json.Marshal(dashboardHistorySignalsData{
		DashboardHistory: dashboardHistorySignalFromAnalytics(data.Analytics),
	})
	return string(b)
}

func dashboardHistorySignalFromAnalytics(data DashboardAnalyticsData) dashboardHistorySignal {
	signal := dashboardHistorySignal{
		CPU:    make([]dashboardCPUHistorySignal, 0, len(data.CPU)),
		Memory: make([]dashboardMemoryHistorySignal, 0, len(data.Memory)),
		Disks:  make([]dashboardDiskHistorySignal, 0, len(data.Disks)),
	}

	for _, sample := range data.CPU {
		signal.CPU = append(signal.CPU, dashboardCPUHistorySignal{
			ObservedAt:     sample.ObservedAt,
			UsedPercent:    sample.UsedPercent,
			CoresLogical:   sample.CoresLogical,
			CoresPhysical:  sample.CoresPhysical,
			PerCorePercent: sample.PerCorePercent,
			Load1M:         sample.Load1M,
			Load5M:         sample.Load5M,
			Load15M:        sample.Load15M,
		})
	}
	for _, sample := range data.Memory {
		signal.Memory = append(signal.Memory, dashboardMemoryHistorySignal{
			ObservedAt:            sample.ObservedAt,
			VirtualUsedPercent:    sample.VirtualUsedPercent,
			VirtualUsedBytes:      sample.VirtualUsedBytes,
			VirtualAvailableBytes: sample.VirtualAvailableBytes,
			VirtualTotalBytes:     sample.VirtualTotalBytes,
			SwapUsedPercent:       sample.SwapUsedPercent,
			SwapUsedBytes:         sample.SwapUsedBytes,
			SwapAvailableBytes:    sample.SwapAvailableBytes,
			SwapTotalBytes:        sample.SwapTotalBytes,
		})
	}
	for _, sample := range data.Disks {
		signal.Disks = append(signal.Disks, dashboardDiskHistorySignal{
			ObservedAt:  sample.ObservedAt,
			IsTotal:     sample.IsTotal,
			Mount:       sample.Mount,
			Device:      sample.Device,
			Filesystem:  sample.Filesystem,
			UsedPercent: sample.UsedPercent,
			UsedBytes:   sample.UsedBytes,
			FreeBytes:   sample.FreeBytes,
			TotalBytes:  sample.TotalBytes,
		})
	}

	return signal
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
