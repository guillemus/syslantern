package server

import (
	"log/slog"
	"net/http"
)

func NewCrossOriginProtection(log *slog.Logger) *http.CrossOriginProtection {
	protection := http.NewCrossOriginProtection()
	protection.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Warn(
			"cross-origin request rejected",
			"path", r.URL.Path,
			"origin", r.Header.Get("Origin"),
			"sec_fetch_site", r.Header.Get("Sec-Fetch-Site"),
		)
		http.Error(w, "forbidden", http.StatusForbidden)
	}))

	return protection
}
