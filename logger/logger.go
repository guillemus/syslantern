package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"

	"app/config"

	"github.com/lmittmann/tint"
)

func NewLogger(cfg config.Config) *slog.Logger {
	w := os.Stderr

	if cfg.Dev {
		opts := &tint.Options{
			Level:     slog.LevelDebug,
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey && len(groups) == 0 {
					return slog.Attr{}
				}
				return a
			},
			TimeFormat: "",
			NoColor:    false,
		}
		return slog.New(tint.NewHandler(w, opts))
	}

	opts := &slog.HandlerOptions{
		Level:       slog.LevelInfo,
		AddSource:   true,
		ReplaceAttr: nil,
	}

	return slog.New(slog.NewJSONHandler(w, opts))
}

func PrettyJSON(s string) string {
	var obj any
	if err := json.Unmarshal([]byte(s), &obj); err != nil {
		return s
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(s), "", "  "); err != nil {
		return s
	}
	return buf.String()
}
