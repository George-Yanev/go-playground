package main

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
	"strings"
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
	allowedCh := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	allowedChSlice := strings.Split(allowedCh, "")
	cache := make(map[string]string)

	// database

	// HTTP server
	postRequest := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		// get the seed

		// get our locally generated characters
		rndChars := generateRandomUrlCharacters(3, allowedChSlice)
		var uFound bool
		for i := 0; i < 3; i++ {
			if _, ok := cache[rndChars]; !ok {
				cache[rndChars] = req.OriginalUrl
				uFound = true
				break
			}
		}

		if !uFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
		}

		rsp := Response{
			ShortCode:   rndChars
			ShortUrl:    "http://short",
			OriginalUrl: req.OriginalUrl,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rsp)
	})

	http.Handle("POST /api/shorten", postRequest)

	log.Fatal(http.ListenAndServe(":5000", nil))
}

func generateRandomUrlCharacters(length int, ch []string) string {
	var b strings.Builder
	for i := 0; i < length; i++ {
		n := rand.IntN(len(ch))
		b.WriteString(ch[n])
	}
	return b.String()
}
