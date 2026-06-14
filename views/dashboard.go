package views

import (
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Dashboard() Node {
	return Div(
		Class("flex h-dvh bg-zinc-950 font-mono text-zinc-100"),
		Data("signals", `{count: 0}`),
		Main(
			Class("flex min-w-0 flex-1 flex-col"),
			Div(
				Class("flex w-full flex-1 flex-col items-start gap-3 overflow-y-auto px-6 py-4"),
				H1(Class("text-3xl font-semibold tracking-tight"), Text("Dashboard")),
				P(Class("text-zinc-400"), Text("You are logged in.")),
				Section(
					Class("w-full max-w-2xl space-y-3 rounded-xl border border-zinc-800 bg-zinc-900 p-6"),
					H2(Class("text-xl font-semibold"), Text("Datastar Example")),
					P(Class("text-zinc-400"), Text("Ask the server to increment a counter and stream the UI update back over SSE.")),
					Button(
						Class("rounded-md bg-orange-600 px-3 py-2 font-medium text-white transition hover:brightness-110"),
						Data("on:click", "@get('/dash/example')"),
						Data("indicator", "loading"),
						Text("Increment On Server"),
					),
					P(
						Class("text-sm text-zinc-500"),
						Data("show", "$loading"),
						StyleAttr("display: none;"),
						Text("Updating..."),
					),
					P(
						Text("Current count: "),
						Strong(
							Data("text", "$count"),
							Text("0"),
						),
					),
					Div(
						ID("dashboard-example-result"),
						Class("rounded-lg border border-zinc-800 bg-zinc-950 p-4 text-zinc-300"),
						P(Text("Press the button to load a Datastar response from the server.")),
					),
				),
				Form(
					Method("POST"),
					Action("/logout"),
					Button(
						Class("rounded-md bg-orange-600 px-3 py-2 font-medium text-white transition hover:brightness-110"),
						Attr("type", "submit"),
						Text("Logout"),
					),
				),
			),
		),
	)
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
