package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Landing() Node {
	return Div(
		Class("flex h-dvh bg-zinc-950 font-mono text-zinc-100"),
		Main(
			Class("flex min-w-0 flex-1 flex-col"),
			Div(
				Class("flex w-full flex-1 flex-col items-start gap-3 overflow-y-auto px-6 py-4"),
				H1(Class("text-3xl font-semibold tracking-tight"), Text("Template")),
				P(Class("text-zinc-400"), Text("Go, Jet, Datastar, SCS, and PostgreSQL.")),
				P(
					Class("text-zinc-400"),
					A(Class("text-orange-400 underline underline-offset-4 hover:text-orange-300"), Attr("href", "/sign-in"), Text("Sign In")),
					Text(" or "),
					A(Class("text-orange-400 underline underline-offset-4 hover:text-orange-300"), Attr("href", "/sign-up"), Text("Sign Up")),
				),
			),
		),
	)
}
