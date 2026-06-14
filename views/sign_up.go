package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type SignUpData struct {
	Email string
	Error string
}

func SignUp(data SignUpData) Node {
	return Div(
		Class("flex h-dvh items-center justify-center bg-zinc-950 px-4 font-mono text-zinc-100"),
		Div(
			Class("w-full max-w-sm rounded-xl border border-zinc-800 bg-zinc-900 p-6"),
			H1(Class("mb-4 text-2xl font-semibold tracking-tight"), Text("Sign Up")),
			ErrorParagraph(data.Error),
			Form(
				Class("space-y-3"),
				Method("POST"),
				Action("/sign-up"),
				Label(
					Class("block text-sm text-zinc-400"),
					For("email"),
					Text("Email"),
				),
				Input(
					Class("w-full rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2 text-zinc-100 outline-none focus:border-orange-500"),
					Type("email"),
					ID("email"),
					Name("email"),
					Value(data.Email),
					Required(),
					AutoFocus(),
				),
				Label(
					Class("block text-sm text-zinc-400"),
					For("password"),
					Text("Password"),
				),
				Input(
					Class("w-full rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2 text-zinc-100 outline-none focus:border-orange-500"),
					Type("password"),
					ID("password"),
					Name("password"),
					Required(),
				),
				Button(
					Class("w-full rounded-md bg-orange-600 px-3 py-2 font-medium text-white transition hover:brightness-110"),
					Attr("type", "submit"),
					Text("Sign Up"),
				),
			),
			P(
				Class("mt-4 text-center text-sm text-zinc-400"),
				Text("Already have an account? "),
				A(Class("text-orange-400 underline underline-offset-4 hover:text-orange-300"), Attr("href", "/sign-in"), Text("Sign in")),
			),
		),
	)
}
