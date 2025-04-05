package main

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
)

type Request struct {
	OriginalUrl string `json:"url"`
}

type Response struct {
	// Request
	ShortCode   string `json:"short_code"`
	ShortUrl    string `json:"short_url"`
	OriginalUrl string `json:"url"`
}

func main() {

	// database

	// HTTP server
	postRequest := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		cache := make(map[int]string)
		rnd := rand.IntN(1_000_000_000)
		_, ok := cache[rnd]
		if !ok {
			cache[rnd] = req.OriginalUrl
		}


		rsp := Response{
			ShortCode:   "abc123",
			ShortUrl:    "http://short",
			OriginalUrl: req.OriginalUrl,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rsp)
	})

	http.Handle("POST /api/shorten", postRequest)

	log.Fatal(http.ListenAndServe(":5000", nil))
}
