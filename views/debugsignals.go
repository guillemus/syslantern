package views

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func DebugSignals(dev bool) Node {
	if !dev {
		return nil
	}

	return Div(
		Class("fixed right-5 bottom-2 z-10 w-[500px] overflow-hidden bg-neutral-950 p-2"),
		Data("signals:debug-signals-collapsed", "true"),
		Div(
			Class("flex items-center justify-between gap-2"),
			Small(Text("signals:")),
			Button(
				Type("button"),
				Class("cursor-pointer rounded border border-neutral-800 bg-transparent px-2 font-[inherit] text-neutral-100"),
				Data("text", "$debugSignalsCollapsed ? 'show' : 'hide'"),
				Data("on:click", `
					$debugSignalsCollapsed = !$debugSignalsCollapsed;
				`),
			),
		),
		Pre(
			Class("overflow-hidden whitespace-pre bg-neutral-950"),
			Data("show", "!$debugSignalsCollapsed"),
			Data("json-signals", ""),
		),
	)
}
