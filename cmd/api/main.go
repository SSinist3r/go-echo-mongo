package main

import (
	"log"

	"go-echo-mongo/internal/server"
)

func main() {
	// Create and start server
	s := server.NewServer()
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
