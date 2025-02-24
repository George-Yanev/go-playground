package main

import (
	"log"
	"os"

	"github.com/George-Yanev/go-fun/internal/urlshortener"
)

func main() {
	db, err := urlshortener.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize the db: %v", err)
	}
	defer db.Close()

	shortUrlHost := os.Getenv("SHORT_URL_HOST")
	if shortUrlHost == "" {
		log.Fatalln("Please setup short_url_host environment variable")
	}

	workCh := make(chan urlshortener.WorkRequest)
	seedCh := make(chan urlshortener.SeedRequest)

	go urlshortener.Manager(db, seedCh)
	urlshortener.StartWorkers(db, workCh, seedCh, 10)
	urlshortener.StartHttpServer(workCh, shortUrlHost)
}
