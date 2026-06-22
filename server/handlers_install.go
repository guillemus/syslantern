package server

import (
	_ "embed"
	"net/http"
	"strings"
)

//go:embed install.sh
var installScript string

func (s *Server) HandleInstallScript(w http.ResponseWriter, r *http.Request) {
	script := strings.ReplaceAll(installScript, "__SYSLANTERN_HUB_URL__", hubURL(r))

	w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	w.Write([]byte(script))
}
