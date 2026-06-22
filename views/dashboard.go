package views

import (
	"fmt"
	"net/url"
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

type DashboardData struct {
	AgentID   string
	Stats     DashboardStatsData
	Analytics DashboardAnalyticsData
}

type AgentsIndexPageData struct {
	Agents         []AgentsIndexData
	InstallCommand string
}

type AgentsIndexData struct {
	ID        string
	Name      string
	Version   string
	UpdatedAt time.Time
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

func (r *Renderer) AgentsIndex(data AgentsIndexPageData) Node {
	return Div(
		Class("min-h-dvh bg-zinc-950 p-6 font-mono text-zinc-100"),
		Main(
			Class("mx-auto max-w-5xl space-y-6"),
			Header(
				Class("flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between"),
				Div(
					H1(Class("text-3xl font-semibold"), Text("Agents")),
					P(Class("mt-2 text-sm text-zinc-500"), Text("Agents available for your account.")),
				),
				Button(
					Class("rounded-md bg-orange-600 px-3 py-2 text-sm font-medium text-white transition hover:brightness-110"),
					Attr("onclick", "agent_install_dialog.showModal()"),
					Text("Add agent"),
				),
			),
			r.DataGet("init", "/events"),
			agentInstallDialog(data.InstallCommand),
			Div(
				Class("overflow-hidden rounded-xl border border-zinc-800 bg-zinc-900"),
				Table(
					Class("w-full text-left text-sm"),
					THead(Tr(
						Th(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text("Name")),
						Th(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text("Agent ID")),
						Th(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text("Version")),
						Th(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text("Updated")),
					)),
					r.agentsTableBody(data.Agents),
				),
			),
		),
	)
}

func (r *Renderer) agentsTableBody(agents []AgentsIndexData) Node {
	rows := make([]Node, 0, len(agents))
	for _, agent := range agents {
		rows = append(rows, r.agentRow(agent))
	}
	if len(rows) == 0 {
		rows = append(rows, Tr(
			Td(ColSpan("4"), Class("p-6 text-zinc-500"), Text("No agents added yet.")),
		))
	}
	nodes := []Node{ID("agents-table-body")}
	nodes = append(nodes, rows...)
	return TBody(nodes...)
}

func agentInstallDialog(command string) Node {
	return Dialog(
		ID("agent_install_dialog"),
		Class("fixed inset-0 m-auto w-[min(42rem,calc(100vw-2rem))] max-h-[calc(100vh-2rem)] rounded-xl border border-zinc-800 bg-zinc-900 p-0 text-zinc-100 shadow-2xl backdrop:bg-black/70"),
		Div(
			Class("p-6"),
			H2(Class("text-xl font-semibold"), Text("Install an agent")),
			P(Class("mt-2 text-sm text-zinc-500"), Text("Run this command on the VPS you want to monitor.")),
			Div(
				Class("mt-4 rounded-lg border border-zinc-800 bg-zinc-950 p-4"),
				Pre(Class("whitespace-pre-wrap break-all text-sm leading-6 text-zinc-100"), Code(Text(command))),
			),
			Form(
				Class("mt-5 flex justify-end"),
				Attr("method", "dialog"),
				Button(Class("rounded-md border border-zinc-700 px-3 py-2 text-sm text-zinc-100 hover:bg-zinc-800"), Text("Close")),
			),
		),
	)
}

func (r *Renderer) agentRow(data AgentsIndexData) Node {
	href := r.URL("GET", "/agents/"+url.PathEscape(data.ID))
	return Tr(
		Td(Class("border-b border-zinc-800 px-4 py-3"), A(Href(href), Class("font-semibold text-zinc-100 hover:text-emerald-300"), Text(valueOr(data.Name, "unnamed")))),
		Td(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text(data.ID)),
		Td(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text(valueOr(data.Version, "—"))),
		Td(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text(updatedAt(data.UpdatedAt))),
	)
}

func (r *Renderer) Dashboard(data DashboardData) Node {
	return Div(
		Class("min-h-dvh bg-zinc-950 p-4 font-mono text-zinc-100 sm:p-6"),
		r.dashboardEventsDataGet(data),
		Main(
			Class("mx-auto flex max-w-7xl flex-col gap-4"),
			dashboardHeader(data),
			DashboardStats(data.Stats),
			DashboardHistory(data.Analytics),
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
