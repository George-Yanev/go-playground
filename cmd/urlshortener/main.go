package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/George-Yanev/go-playground/internal/urlshortener"
)

func main() {
	db, err := urlshortener.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize the db: %v", err)
	}
	defer db.Close()

	err = syncSeedFromUrlMapping(db)
	if err != nil {
		log.Fatalf("Cannot sync seed table from url_mapping. Error: %v", err)
	}

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

func syncSeedFromUrlMapping(db *sql.DB) error {
	u := urlshortener.NewUrlMapping(db)
	seedDb := urlshortener.NewSeedsDb(db)

	seeds, err := seedDb.SelectSeedByStatus(1) // get used seeds 0 - available, 1 - used, 2 - exhausted
	if err != nil {
		return fmt.Errorf("Cannot get Seed by status: %w", err)
	}

	for _, s := range seeds {
		counterUsed, err := u.GetSeedCounter(s.Seed)
		if err != nil {
			return fmt.Errorf("Cannot get url_mapping seed counter: %w", err)
		}

		status := 0
		if counterUsed == s.CounterSize {
			status = 2
		}
		seedDb.SetSeedStatusAndCounter(s.Seed, counterUsed, status)
	}
	return nil
}
