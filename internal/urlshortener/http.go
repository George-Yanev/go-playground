package urlshortener

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func StartHttpServer(workCh chan<- WorkRequest, shortUrlHost string) {
	shortenHandle := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allows all origins
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		var req URLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		fmt.Println("shortUrlHost: " + shortUrlHost)

		doneCh := make(chan WorkResponse, 1)

		work := WorkRequest{
			OriginalUrl:  req.OriginalURL,
			ShortUrlHost: shortUrlHost,
			DoneCh:       doneCh,
		}
		workCh <- work
		resp := <-doneCh
		if resp.Err != nil {
			http.Error(w, fmt.Sprintf("Failed to shorten URL: %v", resp.Err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"short_code":   "abc123",
			"short_url":    resp.ShortUrl,
			"original_url": req.OriginalURL,
		})
	})

	redirectHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})

	http.Handle("/api/shorten", shortenHandle)
	http.Handle("/", redirectHandler)
	log.Fatal(http.ListenAndServe(":5000", nil)) // nil uses the default ServeMux
}
