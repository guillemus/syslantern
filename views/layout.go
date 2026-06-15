package views

import (
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

func (r *Renderer) Layout(title string, body Node) Node {
	return HTML5(HTML5Props{
		Title:    title,
		Language: "en",
		Head: []Node{
			Meta(Attr("charset", "UTF-8")),
			Meta(Attr("name", "viewport"), Attr("content", "width=device-width, initial-scale=1.0")),
			Link(
				Attr("rel", "stylesheet"),
				Attr("href", fmt.Sprintf("%s?v=%s", r.URL("GET", "/public/styles.css"), r.AssetVersion)),
			),
			Script(
				Attr("type", "module"),
				Attr("src", "https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.0-RC.7/bundles/datastar.js"),
			),
		},
		Body: []Node{body},
	})
}
