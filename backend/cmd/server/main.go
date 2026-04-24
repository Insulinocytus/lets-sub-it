package main

import (
	"log"

	"lets-sub-it-api/internal/app"
)

func main() {
	config := app.LoadConfig()
	log.Printf("starting lets-sub-it-api on %s", config.Addr)
}
