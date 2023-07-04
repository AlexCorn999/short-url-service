package main

import (
	"net/http"

	"github.com/AlexCorn999/short-url-service/internal/app/handlers"
)

func main() {
	store := handlers.NewStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/", store.StringAcceptAndBack)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
