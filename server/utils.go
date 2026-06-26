package server

import (
	"io"
	"log/slog"
	"net/http"
	"syslantern/validate"

	"github.com/bytedance/sonic"
)

func readBody(r *http.Request, v any) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if err := sonic.Unmarshal(body, v); err != nil {
		return err
	}

	return validate.V.Struct(v)
}

func writeErr(w http.ResponseWriter, err error, errMsg string) {
	slog.Error("http error", "err", err)
	http.Error(w, errMsg, http.StatusBadRequest)
}

func writeJSON(w http.ResponseWriter, v any) {
	b, err := sonic.Marshal(v)
	if err != nil {
		writeErr(w, err, "failed to write JSON response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(b)
	if err != nil {
		slog.Error("http write error", "err", err)
	}
}

func writeText(w http.ResponseWriter, text string) {
	w.Write([]byte(text))
}

func isDatastarRequest(r *http.Request) bool {
	return r.Header.Get("Datastar-Request") == "true"
}
