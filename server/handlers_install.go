package server

import (
	"net/http"
	"strings"
	"syslantern"
)

func (s *Server) HandleInstallScript(w http.ResponseWriter, r *http.Request) {
	script := strings.ReplaceAll(syslantern.InstallScript, "__SYSLANTERN_HUB_URL__", hubURL(r))

	w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	w.Write([]byte(script))
}
