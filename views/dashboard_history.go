package views

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func DashboardHistory(data DashboardAnalyticsData) Node {
	return Section(
		ID("dashboard-history"),
		Class("contents"),
		Section(
			Class("grid gap-4 xl:grid-cols-[1.5fr_1fr]"),
			cpuHistory(data.CPU),
			memoryPressure(data.Memory),
		),
		diskPressure(data.Disks),
	)
}

func cpuHistory(points []DashboardCPUHistoryData) Node {
	current := latestCPU(points)
	load := "Waiting for CPU history"
	if current != nil {
		load = fmt.Sprintf("load %.2f . %d logical cores", current.Load1M, current.CoresLogical)
	}
	coreRows := append([]Node{Class("core-rows mt-4 space-y-2")}, cpuCoreRows(points)...)

	return Section(
		Class("block min-w-0 border border-zinc-700 bg-zinc-950 p-4"),
		Div(
			Class("flex flex-col gap-1 sm:flex-row sm:items-start sm:justify-between"),
			Div(
				H2(Class("text-lg font-semibold"), Text("CPU history")),
				P(Class("text-sm text-zinc-500"), Text(load)),
			),
			Div(
				Class("flex items-center gap-2 text-xs text-zinc-500"),
				legendItem("bg-emerald-400", "low"),
				legendItem("bg-amber-400", "warm"),
				legendItem("bg-red-400", "hot"),
			),
		),
		Div(
			Class("graph-grid mt-4 h-64 border border-zinc-700 bg-zinc-950"),
			historyCanvas(cpuPointsJSON(points), "UsedPercent", "rgb(52, 211, 153)", "rgba(52, 211, 153, 0.42)", "CPU utilization history"),
		),
		Div(coreRows...),
	)
}

func legendItem(color string, label string) Node {
	return Span(
		Class("inline-flex items-center gap-1"),
		Span(Class("h-2 w-2 "+color)),
		Text(label),
	)
}

func cpuCoreRows(points []DashboardCPUHistoryData) []Node {
	current := latestCPU(points)
	coreCount := 0
	if current != nil {
		coreCount = len(current.PerCorePercent)
		if coreCount == 0 {
			coreCount = current.CoresLogical
		}
	}
	if coreCount == 0 {
		return []Node{Div(Class("text-sm text-zinc-500"), Text("Waiting for per-core samples"))}
	}

	rows := make([]Node, 0, coreCount)
	for core := 0; core < coreCount; core++ {
		values := cpuCoreValues(points, core)
		last := 0.0
		if len(values) > 0 {
			last = values[len(values)-1]
		}
		cells := make([]Node, 0, len(values))
		for _, value := range values {
			cells = append(cells, Span(Class("meter-cell "+meterCellClass(value))))
		}
		meterNodes := append([]Node{Class("flex h-4 gap-px overflow-hidden")}, cells...)
		rows = append(rows, Div(
			Class("grid grid-cols-[3.5rem_minmax(0,1fr)_3rem] items-center gap-2 text-xs"),
			Span(Class("text-zinc-500"), Text(fmt.Sprintf("cpu%d", core))),
			Div(meterNodes...),
			Span(Class("text-right text-zinc-300"), Text(percent(last))),
		))
	}
	return rows
}

func cpuCoreValues(points []DashboardCPUHistoryData, core int) []float64 {
	start := max(len(points)-72, 0)
	values := make([]float64, 0, len(points)-start)
	for _, point := range points[start:] {
		value := point.UsedPercent
		if core < len(point.PerCorePercent) {
			value = point.PerCorePercent[core]
		}
		values = append(values, value)
	}
	return values
}

func meterCellClass(value float64) string {
	if value > 80 {
		return "on-high"
	}
	if value > 60 {
		return "on-mid"
	}
	return "on-low"
}

func memoryPressure(points []DashboardMemoryHistoryData) Node {
	current := latestMemory(points)
	badge := "steady"
	ramText := "Waiting for RAM samples"
	swapText := "Waiting for swap samples"
	ramWidth := 0.0
	swapWidth := 0.0
	if current != nil {
		if current.VirtualUsedPercent >= 85 {
			badge = "hot"
		}
		ramText = fmt.Sprintf("%s used . %s free", formatBytes(current.VirtualUsedBytes), formatBytes(current.VirtualAvailableBytes))
		swapText = fmt.Sprintf("%s used . %s free", formatBytes(current.SwapUsedBytes), formatBytes(current.SwapAvailableBytes))
		ramWidth = current.VirtualUsedPercent
		swapWidth = current.SwapUsedPercent
	}

	return Section(
		Class("block min-w-0 border border-zinc-700 bg-zinc-950 p-4"),
		Div(
			Class("flex items-start justify-between gap-4"),
			Div(
				H2(Class("text-lg font-semibold"), Text("Memory pressure")),
				P(Class("text-sm text-zinc-500"), Text("Used memory and swap pressure.")),
			),
			Span(Class("rounded bg-amber-500 px-2 py-1 text-xs font-semibold text-zinc-950"), Text(badge)),
		),
		Div(
			Class("graph-grid mt-4 h-64 border border-zinc-700 bg-zinc-950"),
			historyCanvas(memoryPointsJSON(points), "VirtualUsedPercent", "rgb(251, 191, 36)", "rgba(251, 191, 36, 0.48)", "Memory usage history"),
		),
		Div(
			Class("mt-4 space-y-3 text-sm"),
			memoryMeter("RAM", ramText, "bg-amber-400", ramWidth),
			memoryMeter("Swap", swapText, "bg-cyan-400", swapWidth),
		),
	)
}

func memoryMeter(label string, text string, color string, width float64) Node {
	return Div(
		Div(
			Class("mb-1 flex justify-between gap-4 text-zinc-400"),
			Span(Text(label)),
			Span(Text(text)),
		),
		Div(
			Class("h-3 overflow-hidden bg-zinc-800"),
			Div(Class("h-full "+color), Attr("style", widthStyle(width))),
		),
	)
}

func diskPressure(points []DashboardDiskHistoryData) Node {
	disks := latestDisks(points)
	badge := "waiting"
	if len(disks) > 0 {
		status := "healthy"
		if disks[0].UsedPercent >= 85 {
			status = "needs attention"
		}
		badge = fmt.Sprintf("%s %s", valueOr(disks[0].Mount, "disk"), status)
	}

	rows := make([]Node, 0, len(disks))
	if len(disks) == 0 {
		rows = append(rows, Tr(Td(Class("px-4 py-6 text-zinc-500"), ColSpan("5"), Text("Waiting for disk samples"))))
	} else {
		for _, disk := range disks {
			rows = append(rows, diskRow(disk))
		}
	}

	return Section(
		Class("block min-w-0 border border-zinc-700 bg-zinc-950 p-4"),
		Div(
			Class("flex flex-col gap-1 sm:flex-row sm:items-start sm:justify-between"),
			Div(
				H2(Class("text-lg font-semibold"), Text("Disk pressure")),
				P(Class("text-sm text-zinc-500"), Text("Sorted by least free space.")),
			),
			Span(Class("rounded bg-red-500 px-2 py-1 text-xs font-semibold text-white"), Text(badge)),
		),
		Div(
			Class("mt-4 overflow-x-auto"),
			Table(
				Class("w-full text-left text-sm"),
				THead(Tr(
					Class("text-zinc-500"),
					Th(Class("border-b border-zinc-800 px-4 py-2"), Text("mount")),
					Th(Class("border-b border-zinc-800 px-4 py-2"), Text("fs")),
					Th(Class("border-b border-zinc-800 px-4 py-2"), Text("used")),
					Th(Class("border-b border-zinc-800 px-4 py-2"), Text("free")),
					Th(Class("border-b border-zinc-800 px-4 py-2"), Text("history")),
				)),
				TBody(rows...),
			),
		),
	)
}

func diskRow(disk DashboardDiskHistoryData) Node {
	tone := "text-emerald-400"
	fill := "bg-emerald-400"
	if disk.UsedPercent >= 85 {
		tone = "text-red-400"
		fill = "bg-red-400"
	} else if disk.UsedPercent >= 65 {
		tone = "text-amber-400"
		fill = "bg-amber-400"
	}

	return Tr(
		Th(Class("border-b border-zinc-900 px-4 py-3 font-semibold"), Text(valueOr(valueOr(disk.Mount, disk.Device), "-"))),
		Td(Class("border-b border-zinc-900 px-4 py-3 text-zinc-300"), Text(valueOr(disk.Filesystem, "-"))),
		Td(Class("border-b border-zinc-900 px-4 py-3 "+tone), Text(percent(disk.UsedPercent))),
		Td(Class("border-b border-zinc-900 px-4 py-3 text-zinc-300"), Text(formatBytes(disk.FreeBytes))),
		Td(
			Class("border-b border-zinc-900 px-4 py-3"),
			Div(
				Class("h-2 w-40 bg-zinc-800"),
				Div(Class("h-full "+fill), Attr("style", widthStyle(disk.UsedPercent))),
			),
		),
	)
}

func historyCanvas(points string, valueKey string, stroke string, fill string, label string) Node {
	return El(
		"syslantern-history-canvas",
		Class("history-canvas"),
		Data("points", points),
		Data("value-key", valueKey),
		Data("stroke", stroke),
		Data("fill", fill),
		Attr("aria-label", label),
	)
}

func latestCPU(points []DashboardCPUHistoryData) *DashboardCPUHistoryData {
	if len(points) == 0 {
		return nil
	}
	return &points[len(points)-1]
}

func latestMemory(points []DashboardMemoryHistoryData) *DashboardMemoryHistoryData {
	if len(points) == 0 {
		return nil
	}
	return &points[len(points)-1]
}

func latestDisks(points []DashboardDiskHistoryData) []DashboardDiskHistoryData {
	byMount := map[string]DashboardDiskHistoryData{}
	for _, point := range points {
		if point.IsTotal {
			continue
		}
		byMount[valueOr(valueOr(point.Mount, point.Device), "disk")] = point
	}

	disks := make([]DashboardDiskHistoryData, 0, len(byMount))
	for _, disk := range byMount {
		disks = append(disks, disk)
	}
	sort.Slice(disks, func(i, j int) bool {
		return 100-disks[i].UsedPercent < 100-disks[j].UsedPercent
	})
	return disks
}

func widthStyle(width float64) string {
	return fmt.Sprintf("width: %.3f%%", clampPercent(width))
}

func clampPercent(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func cpuPointsJSON(points []DashboardCPUHistoryData) string {
	b, err := json.Marshal(points)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func memoryPointsJSON(points []DashboardMemoryHistoryData) string {
	b, err := json.Marshal(points)
	if err != nil {
		return "[]"
	}
	return string(b)
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
