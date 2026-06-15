package views

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/url"

	"github.com/go-chi/chi/v5"

	. "maragu.dev/gomponents"
)

type Renderer struct {
	AssetVersion string
	Logger       *slog.Logger
	Dev          bool
	Routes       chi.Routes
}

func NewRenderer(logger *slog.Logger, assetVersion string, dev bool) *Renderer {
	return &Renderer{Logger: logger, AssetVersion: url.QueryEscape(assetVersion), Dev: dev}
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
	r.RenderPage(w, "Landing", r.Landing())
}

func (r *Renderer) RenderSignIn(w io.Writer, data SignInData) {
	r.RenderPage(w, "Sign In", r.SignIn(data))
}

func (r *Renderer) RenderSignUp(w io.Writer, data SignUpData) {
	r.RenderPage(w, "Sign Up", r.SignUp(data))
}

func (r *Renderer) RenderDashboard(w io.Writer) {
	r.RenderPage(w, "Dashboard", r.Dashboard())
}

func (r *Renderer) RenderDashboardStatsHTML(data DashboardStatsData) (string, error) {
	var body bytes.Buffer
	node := DashboardStats(data)
	if err := node.Render(&body); err != nil {
		return "", fmt.Errorf("render dashboard stats: %w", err)
	}
	return body.String(), nil
}

func (r *Renderer) RenderDashboardExampleResultHTML(data DashboardExampleResultData) (string, error) {
	var body bytes.Buffer
	node := DashboardExampleResult(data)
	if err := node.Render(&body); err != nil {
		return "", fmt.Errorf("render dashboard example result: %w", err)
	}
	return body.String(), nil
}
