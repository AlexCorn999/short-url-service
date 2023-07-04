package main

import (
	"log"
	"net/http"

	"github.com/AlexCorn999/short-url-service/internal/app/handlers"
)

func main() {
	store := handlers.NewStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/", store.StringAcceptAndBack)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
