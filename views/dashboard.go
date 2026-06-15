package views

import (
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type DashboardStatsData struct {
	HostName  string
	UpdatedAt string
	CPU       string
	Memory    string
	Disks     []DashboardDiskData
}

type DashboardDiskData struct {
	Mount string
	Usage string
}

func (r *Renderer) Dashboard() Node {
	return Div(
		Class("min-h-dvh bg-zinc-950 p-6 font-mono text-zinc-100"),
		r.DataGet("init", "/dashboard/events"),
		Main(
			Class("max-w-3xl space-y-6"),
			H1(Class("text-3xl font-semibold"), Text("Dashboard")),
			DashboardStats(DashboardStatsData{}),
		),
	)
}

func DashboardStats(data DashboardStatsData) Node {
	disks := append([]Node{Class("space-y-2")}, diskRows(data.Disks)...)

	return Section(
		ID("dashboard-stats"),
		Class("space-y-4 rounded-lg border border-zinc-800 bg-zinc-900 p-4"),
		P(Class("text-zinc-400"), Text(valueOr(data.UpdatedAt, "Waiting for events..."))),
		Div(
			Class("grid gap-3 sm:grid-cols-3"),
			statBox("Host", valueOr(data.HostName, "—")),
			statBox("CPU", valueOr(data.CPU, "—")),
			statBox("Memory", valueOr(data.Memory, "—")),
		),
		H2(Class("text-xl font-semibold"), Text("Disks")),
		Div(disks...),
	)
}

func statBox(label string, value string) Node {
	return Div(
		Class("rounded border border-zinc-800 bg-zinc-950 p-3"),
		P(Class("text-sm text-zinc-500"), Text(label)),
		P(Class("text-2xl font-semibold"), Text(value)),
	)
}

func diskRows(disks []DashboardDiskData) []Node {
	if len(disks) == 0 {
		return []Node{P(Class("text-zinc-500"), Text("—"))}
	}

	rows := make([]Node, 0, len(disks))
	for _, disk := range disks {
		rows = append(rows, Div(
			Class("flex justify-between rounded border border-zinc-800 bg-zinc-950 p-3"),
			Span(Text(valueOr(disk.Mount, "—"))),
			Strong(Text(disk.Usage)),
		))
	}
	return rows
}

func valueOr(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
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
