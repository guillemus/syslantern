package views

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/starfederation/datastar-go/datastar"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ssePatchSignals(sse *datastar.ServerSentEventGenerator, v any) {
	bs, err := json.Marshal(v)
	if err != nil {
		slog.Error("failed to marshal", "value", v)
	}

	if err := sse.PatchSignals(bs); err != nil {
		slog.Error("sse.PatchSignals error", "err", err)
	}
}

func ssePatchSignal(sse *datastar.ServerSentEventGenerator, signalName string, signalValue any) {
	ssePatchSignals(sse, map[string]any{signalName: signalValue})
}

func ssePatch(sse *datastar.ServerSentEventGenerator, node Node) {
	var sb strings.Builder
	if err := node.Render(&sb); err != nil {
		slog.Error("view render error", "err", err)
		return
	}

	if err := sse.PatchElements(sb.String()); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		slog.Error("sse.PatchElements error", "err", err)
		return
	}
}

func sseExecJS(sse *datastar.ServerSentEventGenerator, script string) {
	// Add block scope so that declared variables are not on global scope.
	s := fmt.Sprintf("{ %s }", script)
	if err := sse.ExecuteScript(s); err != nil {
		slog.Error("sse.ExecuteScript", "err", err)
	}
}

// InlineScript renders an inline script with a block scope, so declared variables
// don't conflict with global scope
func InlineScript(script string) Node {
	return Script(Raw(fmt.Sprintf("{ %s }", script)))
}

func DataBind(signal, val string) Node {
	return Data("bind:"+signal, val)
}

func DataSignals(signal, val string) Node {
	return Data("signals:"+signal, val)
}

func percent(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	value := float64(bytes)
	for _, suffix := range []string{"KB", "MB", "GB", "TB", "PB"} {
		value = value / unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return fmt.Sprintf("%.1f EB", value/unit)
}

func valueOr(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func updatedAt(sentAt time.Time) string {
	if sentAt.IsZero() {
		return "waiting"
	}
	return fmt.Sprintf("updated %s", sentAt.Format(time.RFC3339))
}
