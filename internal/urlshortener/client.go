package urlshortener

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"strconv"
)

type Seed struct {
	Seed    string
	Counter int
}

type Seeds []Seed

type SeedRequest struct {
	Query   string
	ReplyCn chan Seed
}

type WorkRequest struct {
	OriginalUrl string
	DoneChan    chan WorkResult
}

type WorkResult struct {
	ShortUrl string
}

type URLRequest struct {
	OriginalURL string `json:"original_url"`
}

func Manager(reqCh <-chan SeedRequest) {
	for req := range reqCh {
		// db connection to take seed
		// if seed is not available, ask for one
		s := Seed{
			Seed:    "test",
			Counter: 0,
		}

		req.ReplyCn <- s
	}
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
				if seed == (Seed{}) || seed.Counter == 4095 {
					seedCh <- request
					seed = <-responseCh
				}
				// generate short string
				short_url := base64.URLEncoding.EncodeToString([]byte(seed.Seed + strconv.Itoa(seed.Counter)))
				table := UrlMapping{db: db}
				err := table.Create(work.OriginalUrl, short_url, seed.Seed, seed.Counter)
				if err != nil {
					fmt.Printf("Unable to write to url_mapping. Error: %v", err)
				}
				close(work.DoneChan)
			}
		}()
	}
}
