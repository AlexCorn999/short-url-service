package main

import (
	"log"

	"github.com/AlexCorn999/short-url-service/internal/app/apiserver"
)

func main() {
	config := apiserver.NewConfig()
	config.ParseFlags()
	server := apiserver.New(config)

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
