package main

import (
	"log"
	"net/http"

	"lets-sub-it-api/internal/app"
)

func main() {
	config := app.LoadConfig()
	handler, err := app.NewHTTPHandler(config)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("starting lets-sub-it-api on %s", config.Addr)
	log.Fatal(http.ListenAndServe(config.Addr, handler))
}
