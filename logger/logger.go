package logger

import (
	"log/slog"
	"os"

	"github.com/bytedance/sonic"
	"github.com/lmittmann/tint"
)

func NewLogger(debug bool) *slog.Logger {
	w := os.Stderr

	var logger *slog.Logger

	if debug {
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
		logger = slog.New(tint.NewHandler(w, opts))
	} else {
		opts := &slog.HandlerOptions{
			Level:       slog.LevelInfo,
			AddSource:   true,
			ReplaceAttr: nil,
		}

		logger = slog.New(slog.NewJSONHandler(w, opts))
	}

	slog.SetDefault(logger)

	return logger
}

func PrettyJSON(s string) string {
	var obj any
	if err := sonic.UnmarshalString(s, &obj); err != nil {
		return s
	}
	pretty, err := sonic.MarshalIndent(obj, "", "  ")
	if err != nil {
		return s
	}
	return string(pretty)
}
