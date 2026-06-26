package views

import (
	"strings"

	. "maragu.dev/gomponents"
)

func ThreeDotSvg(classes ...string) Node {
	class := ""
	if len(classes) != 0 {
		class = strings.Join(classes, " ")
	}

	return Rawf(`
		<svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 16 16" class="%s">
			<path d="M0 0h16v16H0z" fill="none" />
			<path fill="currentColor" d="M9.5 13a1.5 1.5 0 1 1-3 0a1.5 1.5 0 0 1 3 0m0-5a1.5 1.5 0 1 1-3 0a1.5 1.5 0 0 1 3 0m0-5a1.5 1.5 0 1 1-3 0a1.5 1.5 0 0 1 3 0" />
		</svg>
	`, class)
}
