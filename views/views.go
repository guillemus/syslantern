package views

import (
	"bytes"
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

func (r *Renderer) RenderString(node Node) string {
	var body bytes.Buffer
	if err := node.Render(&body); err != nil {
		r.Logger.Error("render node to string error", "err", err)
	}
	return body.String()
}

func (r *Renderer) RenderPage(w io.Writer, title string, body Node) {
	r.Render(w, r.Layout(title, body))
}
