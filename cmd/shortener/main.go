package main

import (
	"log"

	"github.com/AlexCorn999/short-url-service/internal/app/apiserver"
)

func main() {
	server := apiserver.APIServer{}../store/
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
