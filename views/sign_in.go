package views

import (
	"net/http"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type SignInSignals struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *Renderer) RenderSignInGenericAuthErr(w http.ResponseWriter, email string) {
	r.RenderPage(w, "sign in", r.SignIn(email, "Something went wrong. Please try again."))

}

func (r *Renderer) RenderSignInInvalidCredsErr(w http.ResponseWriter, email string) {
	r.RenderPage(w, "sign in", r.SignIn(email, "Invalid email or password."))
}

func (r *Renderer) RenderSignIn(w http.ResponseWriter) {
	r.RenderPage(w, "sign in", r.SignIn("", ""))
}

func (r *Renderer) SignIn(email string, err string) Node {
	return Div(
		Class("flex h-dvh items-center justify-center bg-zinc-950 px-4 font-mono text-zinc-100"),
		Div(
			Class("w-full max-w-sm rounded-xl border border-zinc-800 bg-zinc-900 p-6"),
			H1(Class("mb-4 text-2xl font-semibold tracking-tight"), Text("Sign In")),
			ErrorParagraph(err),
			Div(
				Class("space-y-3"),
				Label(
					Class("block text-sm text-zinc-400"),
					For("email"),
					Text("Email"),
				),
				Input(
					Class("w-full rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2 text-zinc-100 outline-none focus:border-orange-500"),
					Type("email"),
					ID("email"),
					Data("bind:email", ""),
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
					Data("bind:password", ""),
					Required(),
				),
				Button(
					Class("w-full rounded-md bg-orange-600 px-3 py-2 font-medium text-white transition hover:brightness-110"),
					r.DataPost("on:click", "/sign-in"),
					Text("Sign In"),
				),
			),
			P(
				Class("mt-4 text-center text-sm text-zinc-400"),
				Text("No account? "),
				A(Class("text-orange-400 underline underline-offset-4 hover:text-orange-300"), Attr("href", r.URL("GET", "/sign-up")), Text("Sign up")),
			),
		),
	)
}

func ErrorParagraph(err string) Node {
	if err == "" {
		return Text("")
	}
	return P(
		Class("mb-2 text-sm text-red-400"),
		Text(err),
	)
}
