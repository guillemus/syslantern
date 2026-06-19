package main

import (
	"syslantern/server"
)

func main() {
	srv := server.NewServer()
	srv.Start()
}
