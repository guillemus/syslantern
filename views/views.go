package views

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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

func (r *Renderer) RenderString(node Node) string {
	var body bytes.Buffer
	if err := node.Render(&body); err != nil {
		r.Logger.Error("render node to string error", "err", err)
	}
	return body.String()
}

func (r *Renderer) RenderPage(w http.ResponseWriter, title string, body Node) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	r.Render(w, r.Layout(title, body))
}

func (r *Renderer) RenderAgentsIndex(w http.ResponseWriter, data AgentsIndexPageData) {
	r.RenderPage(w, "Agents", r.AgentsIndex(data))
}

func (r *Renderer) RenderDashboard(w http.ResponseWriter, data DashboardData) {
	r.RenderPage(w, "Dashboard", r.Dashboard(data))
}

func (r *Renderer) RenderDashboardStatsHTML(data DashboardStatsData) string {
	return r.RenderString(DashboardStats(data))
}

func (r *Renderer) RenderDashboardHistoryHTML(data DashboardAnalyticsData) string {
	return r.RenderString(DashboardHistory(data))
}

func (r *Renderer) RenderDashboardExampleResultHTML(data DashboardExampleResultData) (string, error) {
	var body bytes.Buffer
	node := DashboardExampleResult(data)
	if err := node.Render(&body); err != nil {
		return "", fmt.Errorf("render dashboard example result: %w", err)
	}
	return body.String(), nil
}
