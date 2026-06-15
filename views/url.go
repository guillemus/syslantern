package views

import (
	"fmt"

	"github.com/go-chi/chi/v5"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func (r *Renderer) MatchPath(method, path string) {
	if r.Dev {
		ctx := chi.NewRouteContext()
		matches := chi.Routes.Match(r.Routes, ctx, method, path)
		if !matches {
			panic(fmt.Sprintf("method: %s, path: %s is not registered", method, path))
		}
	}
}

func (r *Renderer) URL(method, path string) string {
	r.MatchPath(method, path)
	return path
}

func (r *Renderer) DataGet(name, path string) Node {
	r.MatchPath("GET", path)
	return Data(name, fmt.Sprintf("@get(%q)", path))
}

func (r *Renderer) DataPost(name, path string) Node {
	r.MatchPath("POST", path)
	return Data(name, fmt.Sprintf("@post(%q)", path))
}
