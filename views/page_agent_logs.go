package views

import (
	"fmt"
	"io"
	"time"

	"github.com/starfederation/datastar-go/datastar"
	. "maragu.dev/gomponents"

	. "maragu.dev/gomponents/html"
)

const (
	agentLogsPageID = "agent_logs_page"
	agentLogsID     = "agent_logs"
)

type AgentLogsPageData struct {
	AgentPageData
	Logs []AgentLogEntryData
}

type AgentLogEntryData struct {
	ObservedAt string
	Message    string
}

func (r *Renderer) RenderAgentLogsPage(w io.Writer, data AgentLogsPageData) {
	r.RenderPage(w, valueOr(data.Name, "agent")+" logs", r.AgentLogsPage(data))
}

func (r *Renderer) AgentLogsPage(data AgentLogsPageData) Node {
	return Div(
		ID(agentLogsPageID),
		Data("init", r.Get("/agents/"+data.ID+"/logs/events")),
		Class("h-dvh overflow-hidden bg-zinc-950 p-6 font-mono text-zinc-100"),
		Main(
			Class("mx-auto flex h-full min-h-0 flex-col gap-6"),
			r.agentPageHeader(data.AgentPageData),
			r.agentLogsNav(data.ID),
			r.agentLogs(data.Logs),
			DeleteAgentDialog(),
		),
	)
}

func (r *Renderer) agentLogsNav(agentID string) Node {
	return Div(
		Class("tabs tabs-border"),
		Attr("role", "tablist"),
		A(Href(r.URL("GET", "/agents/"+agentID)), Class("tab text-zinc-400"), Attr("role", "tab"), Text("Metrics")),
		A(Href(r.URL("GET", "/agents/"+agentID+"/logs")), Class("tab tab-active text-zinc-100"), Attr("role", "tab"), Text("Logs")),
	)
}

func (r *Renderer) PatchAgentLogs(sse *datastar.ServerSentEventGenerator, logs []AgentLogEntryData) {
	ssePatch(sse, r.agentLogs(logs))
}

func (r *Renderer) agentLogs(logs []AgentLogEntryData) Node {
	rows := make([]Node, 0, len(logs))
	for _, log := range logs {
		fmt.Println(log.Message)
		rows = append(rows, Tr(
			Td(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500 whitespace-nowrap"), Text(logTimestamp(log.ObservedAt))),
			Td(Class("border-b border-zinc-800 px-4 py-3 text-zinc-100"), Text(log.Message)),
		))
	}
	if len(rows) == 0 {
		rows = append(rows, Tr(Td(ColSpan("2"), Class("p-6 text-zinc-500"), Text("No logs received yet."))))
	}

	return Section(
		ID(agentLogsID),
		Class("logs-scroll min-h-0 flex-1 overflow-y-auto overflow-x-hidden rounded-xl border border-zinc-800 bg-zinc-900 [scrollbar-gutter:stable]"),
		Table(
			Class("w-full text-left text-sm"),
			THead(Class("sticky top-0 z-10 bg-zinc-900 text-zinc-500"), Tr(
				Th(Class("px-4 py-3 font-medium"), Text("Time")),
				Th(Class("px-4 py-3 font-medium"), Text("Message")),
			)),
			TBody(rows...),
		),
	)
}

func logTimestamp(value string) string {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return value
	}
	return parsed.Format("Jan 02, 15:04:05")
}
