package views

import (
	"fmt"
	"log"

	"github.com/go-chi/chi/v5"
)

func (r *Renderer) MatchPath(method, path string) {
	if r.Dev {
		ctx := chi.NewRouteContext()
		matches := chi.Routes.Match(r.Routes, ctx, method, path)
		if !matches {
			log.Fatalf("\033[31mmethod %s, path: %s is not registered\033[0m", method, path)
		}
	}
}

func (r *Renderer) URL(method, path string) string {
	r.MatchPath(method, path)
	return path
}

// Get asserts and creates a datastar @get action.
func (r *Renderer) Get(path string) string {
	r.MatchPath("GET", path)
	return fmt.Sprintf("@get(%q)", path)
}

// Post asserts and creates a datastar @post action.
func (r *Renderer) Post(path string) string {
	r.MatchPath("POST", path)
	return fmt.Sprintf("@post(%q)", path)
}
