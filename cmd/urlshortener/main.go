package main

import (
	"log"

	"github.com/George-Yanev/go-fun/internal/urlshortener"
)

func main() {
	db, err := urlshortener.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize the db: %v", err)
	}
	defer db.Close()

	workCh := make(chan urlshortener.WorkRequest)
	seedCh := make(chan urlshortener.SeedRequest)

	go urlshortener.Manager(seedCh)
	urlshortener.StartWorkers(db, workCh, seedCh, 10)
	urlshortener.StartHttpServer(workCh)
}
