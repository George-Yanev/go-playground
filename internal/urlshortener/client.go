package urlshortener

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"

	"github.com/google/uuid"
)

type Seed struct {
	Seed        string
	CounterUsed int
	CounterSize int
}

type Seeds []Seed

type SeedRequest struct {
	Query   string
	ReplyCn chan Seed
}

type WorkRequest struct {
	OriginalUrl  string
	ShortUrlHost string
	DoneCh       chan<- WorkResponse
}

type WorkResponse struct {
	ShortUrl string
	Err      error
}

type URLRequest struct {
	OriginalURL string `json:"original_url"`
}

func Manager(db *sql.DB, reqCh <-chan SeedRequest) {
	seedsDb := NewSeedsDb(db)

	for req := range reqCh {
		seed, err := seedsDb.Acquire(req.Query)
		if err != nil {
			log.Printf("Error acquiring seed: %v\n", err)
		}

		log.Printf("Client: %s acquired Seed: %v\n", req.Query, seed)
		req.ReplyCn <- seed

		// house keeping. Close a seed if a client has already used one
	}
}

func StartWorkers(db *sql.DB, workCh <-chan WorkRequest, seedCh chan<- SeedRequest, numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go func() {
			var seed Seed
			leaseHolderID := uuid.New().String()

			responseCh := make(chan Seed)
			request := SeedRequest{
				Query:   leaseHolderID,
				ReplyCn: responseCh,
			}

			for work := range workCh {
				if seed == (Seed{}) || seed.CounterUsed == seed.CounterSize {
					seedCh <- request
					seed = <-responseCh
				}
				// generate short string
				cUsed := seed.CounterUsed + 1
				sUrlPath := base64.URLEncoding.EncodeToString([]byte(seed.Seed + strconv.Itoa(cUsed)))
				sUrl := fmt.Sprintf("https://%s/%s", work.ShortUrlHost, sUrlPath)
				um := UrlMapping{db: db}
				err := um.Create(work.OriginalUrl, sUrl, seed.Seed, cUsed)
				if err != nil {
					fmt.Printf("Unable to write to url_mapping. Error: %v\n", err)
				} else {
					seed.CounterUsed = cUsed
				}

				// finish the response regardless of the status
				wr := WorkResponse{
					ShortUrl: sUrl,
					Err:      err,
				}
				work.DoneCh <- wr
				close(work.DoneCh)
			}
		}()
	}
}
