package urlshortener

import (
	"database/sql"
	"encoding/base64"
)

type Seed struct {
	seed    string
	counter int
}

type SeedRequest struct {
	Query   string
	ReplyCn chan Seed
}

type WorkRequest struct {
	OriginalUrl string
	DoneChan    chan struct{}
}

type URLRequest struct {
	OriginalURL string `json:"original_url"`
}

func Manager(reqCh <-chan SeedRequest) {
	req := <-reqCh

	// db connection to take seed
	// if seed is not available, ask for one
	s := Seed{
		seed:    "test",
		counter: 0,
	}

	req.ReplyCn <- s

	// check if we have enough seeds
}

func StartWorkers(db *sql.DB, workCh <-chan WorkRequest, seedCh chan<- SeedRequest, numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go func() {
			var seed Seed

			responseCh := make(chan Seed)
			request := SeedRequest{
				Query:   "test",
				ReplyCn: responseCh,
			}

			for work := range workCh {
				if seed == (Seed{}) || seed.counter == 4095 {
					seedCh <- request
					seed = <-responseCh
				}
				// generate short string
				encoded := base64.NewEncoding(seed.seed + string(seed.counter))
				processWork(encoded)
				close(work.DoneChan)
			}
		}()
	}
}
