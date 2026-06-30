package views

import (
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/starfederation/datastar-go/datastar"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const (
	agentPageID       = "agent_page"
	agentOverviewID   = "agent_overview"
	agentMetricsTabID = "agent_metrics_tab"
)

type AgentPageData struct {
	ID             string
	Status         string
	Name           string
	Version        string
	HostID         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	InstallCommand string
	Metrics        DashboardMetricsData
}

type DashboardMetricsData struct {
	Stats     DashboardStatsData
	Analytics DashboardAnalyticsData
}

func (r *Renderer) RenderAgentPage(w io.Writer, data AgentPageData) {
	r.RenderPage(w, valueOr(data.Name, "agent"), r.AgentPage(data))
}

func (r *Renderer) PatchAgentMetrics(sse *datastar.ServerSentEventGenerator, data DashboardMetricsData) {
	ssePatch(sse, r.agentMetricsDashboard(data))
}

func (r *Renderer) AgentPage(data AgentPageData) Node {
	return MainPageLayout(
		ID(agentPageID),
		Data("init", r.Get("/agents/"+url.PathEscape(data.ID)+"/events")),
		r.agentPageHeader(data),
		r.agentTabs(data),
		DeleteAgentDialog(),
	)
}

func (r *Renderer) agentPageHeader(data AgentPageData) Node {
	agentID := url.PathEscape(data.ID)
	deleteHref := r.URL("POST", "/agents/"+agentID+"/delete")

	return Header(
		Class("flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between"),
		Div(
			A(
				Href(r.URL("GET", "/")),
				Class("text-sm text-zinc-500 hover:text-emerald-300"),
				Text("← Agents"),
			),
			Div(
				Class("mt-3 flex items-center gap-3"),
				agentStatusDot(data.Status),
				H1(Class("text-3xl font-semibold"), Text(valueOr(data.Name, "unnamed"))),
			),
			P(Class("mt-2 text-sm text-zinc-500"), Text(data.ID)),
		),
		Div(
			Class("flex flex-wrap gap-2"),
			Button(
				Type("button"),
				Data("on:click", copyButtonJS(data.InstallCommand)),
				Class("rounded-md border border-zinc-700 px-3 py-2 text-sm text-zinc-100 hover:bg-zinc-800"),
				Text("Copy install command"),
			),
			Button(
				Type("button"),
				Class("rounded-md bg-red-700 px-3 py-2 text-sm font-medium text-white transition hover:brightness-110"),
				Data("on:click", `
					$`+deleteAgentNameSignal+` = `+strconv.Quote(valueOr(data.Name, "unnamed"))+`
					$`+deleteAgentURLSignal+` = `+strconv.Quote(deleteHref)+`
					$`+_deleteAgentLoadingSignal+` = false
					`+deleteAgentDialogID+`.showModal()
				`),
				Text("Delete agent"),
			),
		),
	)
}

func (r *Renderer) agentOverview(data AgentPageData) Node {
	return Section(
		ID(agentOverviewID),
		Class("rounded-xl border border-zinc-800 bg-zinc-900 px-4 py-3"),
		Div(
			Class("grid gap-3 text-sm sm:grid-cols-2 lg:grid-cols-3"),
			agentDetail("Status", valueOr(data.Status, "unknown")),
			agentDetail("Version", valueOr(data.Version, "—")),
			agentDetail("Host", valueOr(data.HostID, "waiting for install")),
			agentDetail("Created", timestamp(data.CreatedAt)),
			agentDetail("Updated", timestamp(data.UpdatedAt)),
			agentDetail("ID", data.ID),
		),
	)
}

func (r *Renderer) agentTabs(data AgentPageData) Node {
	agentID := url.PathEscape(data.ID)
	return Div(
		Class("space-y-4"),
		Div(
			Class("tabs tabs-border"),
			Attr("role", "tablist"),
			A(Href(r.URL("GET", "/agents/"+agentID)), Class("tab tab-active text-zinc-100"), Attr("role", "tab"), Text("Metrics")),
			A(Href(r.URL("GET", "/agents/"+agentID+"/logs")), Class("tab text-zinc-400"), Attr("role", "tab"), Text("Logs")),
		),
		r.agentMetricsTab(data),
	)
}

func (r *Renderer) agentMetricsTab(data AgentPageData) Node {
	return Div(
		Class("space-y-4"),
		r.agentOverview(data),
		r.agentMetricsDashboard(data.Metrics),
	)
}

func (r *Renderer) agentMetricsDashboard(data DashboardMetricsData) Node {
	return Div(
		ID(agentMetricsTabID),
		Class("space-y-4"),
		DashboardStats(data.Stats),
		DashboardHistory(data.Analytics),
	)
}

func agentDetail(label string, value string) Node {
	return Div(
		Class("min-w-0"),
		Span(Class("text-zinc-500"), Text(label), Text(": ")),
		Span(Class("break-all text-zinc-100"), Text(valueOr(value, "—"))),
	)
}

func timestamp(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.Format(time.RFC3339)
}

func copyButtonJS(value string) string {
	return `
		navigator.clipboard.writeText(` + strconv.Quote(value) + `);
		const previousText = el.textContent;
		el.textContent = 'Copied!';
		setTimeout(() => {
			el.textContent = previousText;
		}, 1000);
	`
}
