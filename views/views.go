package views

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/url"

	. "maragu.dev/gomponents"
)

type Renderer struct {
	AssetVersion string
	Logger       *slog.Logger
}

func NewRenderer(logger *slog.Logger, assetVersion string) *Renderer {
	return &Renderer{Logger: logger, AssetVersion: url.QueryEscape(assetVersion)}
}

func (r *Renderer) Render(w io.Writer, node Node) {
	if err := node.Render(w); err != nil {
		r.Logger.Error("layout render: render node", "err", err)
	}
}

func (r *Renderer) RenderPage(w io.Writer, title string, body Node) {
	r.Render(w, r.Layout(title, body))
}

func (r *Renderer) RenderLanding(w io.Writer) {
	r.RenderPage(w, "Landing", Landing())
}

func (r *Renderer) RenderSignIn(w io.Writer, data SignInData) {
	r.RenderPage(w, "Sign In", SignIn(data))
}

func (r *Renderer) RenderSignUp(w io.Writer, data SignUpData) {
	r.RenderPage(w, "Sign Up", SignUp(data))
}

func (r *Renderer) RenderDashboard(w io.Writer) {
	r.RenderPage(w, "Dashboard", Dashboard())
}

func (r *Renderer) RenderDashboardExampleResultHTML(data DashboardExampleResultData) (string, error) {
	var body bytes.Buffer
	node := DashboardExampleResult(data)
	if err := node.Render(&body); err != nil {
		return "", fmt.Errorf("render dashboard example result: %w", err)
	}
	return body.String(), nil
}
