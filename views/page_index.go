package views

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/starfederation/datastar-go/datastar"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type AgentsIndexPageData struct {
	Agents []AgentRow
}

type AgentRow struct {
	ID        string
	Name      string
	Version   string
	UpdatedAt time.Time
}

func (r *Renderer) RenderIndexPage(w io.Writer, data AgentsIndexPageData) {
	r.RenderPage(w, "syslantern", r.AgentsIndex(data))
}

func (r *Renderer) PatchIndexPageTableData(sse *datastar.ServerSentEventGenerator, data AgentsIndexPageData) {
	ssePatch(sse, r.AgentsIndex(data))
}

func (r *Renderer) AgentsIndex(data AgentsIndexPageData) Node {
	return Div(
		ID("page"),
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
					Data("on:click", "new_agent_dialog.showModal()"),
					Text("Add agent"),
				),
			),
			r.DataGet("init", "/events"),
			r.AgentsIndexTable(data),
			r.NewAgentDialogForm(),
			r.CopyCommandDialog(""),
		),
	)
}

func (r *Renderer) PatchAgentsIndexTable(
	sse *datastar.ServerSentEventGenerator, data AgentsIndexPageData,
) {
	ssePatchSignal(sse, "_new_agent_loading", false)

	// the index page has a table, if we fat morph we break html dialog state
	ssePatch(sse, r.AgentsIndexTable(data))
}

func (r *Renderer) AgentsIndexTable(data AgentsIndexPageData) Node {
	return Div(
		ID("agents_index_table"),
		Class("overflow-hidden rounded-xl border border-zinc-800 bg-zinc-900"),
		Table(
			Class("w-full text-left text-sm"),
			THead(Tr(
				Th(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text("Name")),
				Th(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text("Agent ID")),
				Th(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text("Version")),
				Th(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text("Updated")),
				Th(Class("border-b border-zinc-800 px-4 py-3 text-right text-zinc-500"), Text("")),
			)),
			r.agentsTableBody(data.Agents),
		),
	)
}

type NewAgentDialogSignals struct {
	NewAgentName string `json:"new_agent_name"`
}

func (r *Renderer) PatchNewAgentDialogErr(w http.ResponseWriter, rr *http.Request, err string) {
	sse := datastar.NewSSE(w, rr)
	ssePatchSignal(sse, "_new_agent_loading", false)
	ssePatch(sse, Div(
		ID("new-agent-dialog-error"),
		Class("mt-3 text-sm text-red-400"),
		P(Text(err)),
	))
}

func (r *Renderer) PatchNewAgentDialogWithCopyCommmand(
	w http.ResponseWriter, rr *http.Request, commandToCopy string,
) {
	sse := datastar.NewSSE(w, rr)
	ssePatch(sse, r.CopyCommandDialog(commandToCopy))
	ssePatchSignals(sse, map[string]any{
		"_new_agent_loading": false,
		"new_agent_name":     "",
	})
	sseExecJS(sse, `
		copy_command_dialog.showModal();
		new_agent_dialog.close();
	`)
}

func (r *Renderer) NewAgentDialogForm() Node {
	return Dialog(
		ID("new_agent_dialog"),
		Class("fixed inset-0 m-auto w-[min(42rem,calc(100vw-2rem))] max-h-[calc(100vh-2rem)] rounded-xl border border-zinc-800 bg-zinc-900 p-0 text-zinc-100 shadow-2xl backdrop:bg-black/70"),
		Form(
			Data("signals:_new_agent_loading", "false"),
			Data("on:submit", fmt.Sprintf(`
				evt.preventDefault()
				$_new_agent_loading = true;
				@post('%s');
			`, r.URL("POST", "/agents/new"))),
			Class("p-6"),
			H2(Class("text-xl font-semibold"), Text("Add agent")),
			P(Class("mt-2 text-sm text-zinc-500"), Text("Give your new agent a name.")),
			Div(
				Class("mt-4 space-y-4"),
				Label(
					Class("block text-sm text-zinc-400"),
					Text("Name"),
					Input(
						Data("bind:new_agent_name", ""),
						Type("text"),
						Name("name"),
						Required(),
						Placeholder("my-vps"),
						Class("mt-1 w-full rounded-md border border-zinc-700 bg-zinc-950 px-3 py-2 text-zinc-100 placeholder-zinc-600 focus:border-emerald-500 focus:outline-none"),
					),
				),
				Div(ID("new-agent-dialog-error")),
				Div(
					Class("flex justify-end gap-2"),
					Button(
						Type("button"),
						Attr("data-on:click", "new_agent_dialog.close()"),
						Class("rounded-md border border-zinc-700 px-3 py-2 text-sm text-zinc-100 hover:bg-zinc-800"),
						Text("Close"),
					),
					Button(
						Type("submit"),
						Class("rounded-md bg-orange-600 px-3 py-2 text-sm font-medium text-white transition hover:brightness-110"),
						Span(
							Data("show", "$_new_agent_loading"),
							Class("loading loading-spinner loading-sm"),
						),
						Span(
							Data("show", "!$_new_agent_loading"),
							Text("Add"),
						),
					),
				),
			),
		),
	)
}

func (r *Renderer) CopyCommandDialog(commandToCopy string) Node {
	return Dialog(
		ID("copy_command_dialog"),
		Class("fixed inset-0 m-auto w-[min(42rem,calc(100vw-2rem))] max-h-[calc(100vh-2rem)] rounded-xl border border-zinc-800 bg-zinc-900 p-0 text-zinc-100 shadow-2xl backdrop:bg-black/70"),
		Div(
			Class("p-6"),
			H2(Class("text-xl font-semibold"), Text("Install agent")),
			P(Class("mt-2 text-sm text-zinc-500"), Text("Run this command on your server to install the agent.")),
			Pre(
				Class("mt-4 overflow-x-auto rounded-md border border-zinc-800 bg-zinc-950 p-4 text-sm text-zinc-100"),
				Code(Text(commandToCopy)),
			),
			Div(
				Class("mt-4 flex justify-end gap-2"),
				Button(
					Type("button"),
					Data("on:click", "copy_command_dialog.close()"),
					Class("rounded-md border border-zinc-700 px-3 py-2 text-sm text-zinc-100 hover:bg-zinc-800"),
					Text("Close"),
				),
				Button(
					Type("button"),
					Data("on:click", fmt.Sprintf("navigator.clipboard.writeText(%s)", commandToCopy)),
					Class("rounded-md bg-orange-600 px-3 py-2 text-sm font-medium text-white transition hover:brightness-110"),
					Text("Copy"),
				),
			),
		),
	)
}

func (r *Renderer) agentsTableBody(agents []AgentRow) Node {
	rows := make([]Node, 0, len(agents))
	for _, agent := range agents {
		rows = append(rows, r.agentRow(agent))
	}
	if len(rows) == 0 {
		rows = append(rows, Tr(
			Td(ColSpan("5"), Class("p-6 text-zinc-500"), Text("No agents added yet.")),
		))
	}
	nodes := []Node{ID("agents-table-body")}
	nodes = append(nodes, rows...)
	return TBody(nodes...)
}

func (r *Renderer) agentRow(data AgentRow) Node {
	agentID := url.PathEscape(data.ID)
	href := r.URL("GET", "/agents/"+agentID)
	deleteHref := r.URL("POST", "/agents/"+agentID+"/delete")

	return Tr(
		Td(Class("border-b border-zinc-800 px-4 py-3"), A(Href(href), Class("font-semibold text-zinc-100 hover:text-emerald-300"), Text(valueOr(data.Name, "unnamed")))),
		Td(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text(data.ID)),
		Td(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text(valueOr(data.Version, "—"))),
		Td(Class("border-b border-zinc-800 px-4 py-3 text-zinc-500"), Text(updatedAt(data.UpdatedAt))),
		Td(
			Class("border-b border-zinc-800 px-4 py-3 text-right"),
			Div(
				Class("dropdown dropdown-end"),
				Button(
					Type("button"),
					Class("btn btn-ghost btn-circle text-zinc-100 hover:bg-zinc-800"),
					TabIndex("0"),
					Role("button"),
					Aria("label", "Agent actions"),
					ThreeDotSvg("text-lg"),
				),
				Ul(
					TabIndex("-1"),
					Class("dropdown-content menu menu-sm rounded-box z-10 mt-2 w-40 border border-zinc-800 bg-zinc-950 p-2 text-zinc-100 shadow-xl shadow-black/40"),
					Li(
						A(Href(href), Class("hover:bg-zinc-800"), Text("View")),
					),
					Li(
						Button(
							Type("button"),
							Class("text-red-400 hover:bg-zinc-800 hover:text-red-300"),
							Data("on:click", fmt.Sprintf(`
								if (confirm('Delete this agent?')) {
									el.querySelector('.x-text').classList.add('hidden')
									el.querySelector('.x-spinner').classList.remove('hidden')
									@post('%s')
								}
							`, deleteHref)),

							Span(Class("x-text"), Text("Delete")),
							Span(Class("x-spinner loading loading-spinner loading-sm hidden")),
						),
					),
				),
			),
		),
	)
}
