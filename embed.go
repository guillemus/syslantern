package app

import (
	"app/config"
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed public/*
var publicFS embed.FS

func GetPublicHandler(cfg config.Config) http.Handler {
	embeddedPublic, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatalf("sub public fs: %v", err)
	}
	publicAssets := http.FileServerFS(embeddedPublic)
	if cfg.Dev {
		publicAssets = http.FileServer(http.Dir("public"))
	}

	return http.StripPrefix("/public/", publicAssets)
}
