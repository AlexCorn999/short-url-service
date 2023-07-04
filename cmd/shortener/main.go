package main

import (
	"log"
	"net/http"

	"github.com/AlexCorn999/short-url-service/internal/app/handlers"
)

const host = ":8080"

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.StringAcceptAndBack)

	log.Fatal(http.ListenAndServe(host, mux))
}
