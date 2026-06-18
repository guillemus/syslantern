package scripts

import (
	"fmt"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func CPUHistory(signal string) Node {
	return historyElement("openlogs-cpu-history", signal, "block min-w-0 border border-zinc-700 bg-zinc-950 p-4")
}

func MemoryPressure(signal string) Node {
	return historyElement("openlogs-memory-pressure", signal, "block min-w-0 border border-zinc-700 bg-zinc-950 p-4")
}

func DiskPressure(signal string) Node {
	return historyElement("openlogs-disk-pressure", signal, "block min-w-0 border border-zinc-700 bg-zinc-950 p-4")
}

func historyElement(tag string, signal string, class string) Node {
	return El(
		tag,
		Class(class),
		Data("attr:data-points", fmt.Sprintf("JSON.stringify(%s)", signal)),
	)
}
