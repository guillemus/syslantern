package main

import (
	"app/server"
)

func main() {
	srv := server.NewServer()
	srv.Start()
}
