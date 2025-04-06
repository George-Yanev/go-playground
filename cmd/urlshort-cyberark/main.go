package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strings"
)

type LocalResponse struct {
	Url  string
	Seed string
	Err  error
}

type Request struct {
	OriginalUrl string `json:"url"`
	ReplyCh     chan LocalResponse
}

type Response struct {
	// Request
	ShortCode   string `json:"short_code"`
	ShortUrl    string `json:"short_url"`
	OriginalUrl string `json:"url"`
}

const (
	BufferedItems   = 100
	LocalSeedLength = 5
)

var cache = make(map[string]string, BufferedItems)
var allowed = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
var allowedSlice = strings.Split(allowed, "")

func main() {
	seedCh := make(chan Request, 100)

	go func(length int) {
		var lResponse LocalResponse
		// get remote seed
		// generate some local seeds
		for req := range seedCh {
			seed, err := bufferedRandomUrlCharacters(req.OriginalUrl)
			if err != nil {
				lResponse.Err = err
			}
			lResponse.Seed = seed
			lResponse.Url = req.OriginalUrl
			req.ReplyCh <- lResponse
			close(req.ReplyCh)
		}
	}(5)

	// HTTP server
	postRequest := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		// Validation logic
		if req.OriginalUrl == "" {
			http.Error(w, "field 'url' is mandatory and must be provided", http.StatusBadRequest)
			return
		}

		// get the seed
		respCh := make(chan LocalResponse, 1)
		req.ReplyCh = respCh
		seedCh <- req
		// wait
		lRsp := <-respCh
		fmt.Println(lRsp)

		w.Header().Set("Content-Type", "application/json")
		if lRsp.Err != nil {
			http.Error(w, "unable to provide short url", http.StatusInternalServerError)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		rsp := Response{
			ShortCode:   lRsp.Seed,
			ShortUrl:    "http://localhost:5000/" + lRsp.Seed,
			OriginalUrl: req.OriginalUrl,
		}

		json.NewEncoder(w).Encode(rsp)
	})

	http.Handle("POST /api/shorten", postRequest)

	log.Fatal(http.ListenAndServe(":5000", nil))
}

func generateRandomUrlCharacters() string {
	var b strings.Builder
	for i := 0; i < LocalSeedLength; i++ {
		n := rand.IntN(len(allowedSlice))
		b.WriteString(allowedSlice[n])
	}
	return b.String()
}

func bufferedRandomUrlCharacters(url string) (string, error) {
	if len(cache) <= BufferedItems/2 {
		for i := 0; i < BufferedItems; i++ {
			rnd := generateRandomUrlCharacters()
			if _, ok := cache[rnd]; !ok {
				cache[rnd] = url
			}
		}
	}

	if len(cache) == 0 {
		return "", errors.New("no random item to return")
	}

	var key string
	for k := range cache {
		key = k
		delete(cache, key)
		break
	}

	return key, nil
}
