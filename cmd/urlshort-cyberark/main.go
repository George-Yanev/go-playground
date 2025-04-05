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

type LocalRequest struct {
	Url     string
	ReplyCh chan LocalResponse
}

type LocalResponse struct {
	Url  string
	Seed string
	Err  error
}

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
	seedCh := make(chan LocalRequest, 100)

	go func(length int) {
		err := errors.New("Unable to generate random seed")
		var lResponse LocalResponse
		// get remote seed
		// generate some local seeds
		for req := range seedCh {
			var uFound bool
			var rndChars string
			for i := 0; i < 3; i++ {
				rndChars = generateRandomUrlCharacters(length, allowedChSlice)
				if _, ok := cache[rndChars]; !ok {
					cache[rndChars] = req.Url
					uFound = true
					break
				}
			}
			if !uFound {
				lResponse.Err = err
			}
			lResponse.Seed = rndChars
			lResponse.Url = req.Url
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

		// get the seed
		respCh := make(chan LocalResponse, 1)
		localReq := LocalRequest{
			Url:     req.OriginalUrl,
			ReplyCh: respCh,
		}
		seedCh <- localReq
		// wait
		lRsp := <-respCh
		fmt.Println(lRsp)

		w.Header().Set("Content-Type", "application/json")
		if lRsp.Err != nil {
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

func generateRandomUrlCharacters(length int, ch []string) string {
	var b strings.Builder
	for i := 0; i < length; i++ {
		n := rand.IntN(len(ch))
		b.WriteString(ch[n])
	}
	return b.String()
}
