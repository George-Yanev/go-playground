package urlshortener

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func StartHttpServer(workCh chan<- WorkRequest, shortUrlHost string) {
	http.HandleFunc("POST /short", func(w http.ResponseWriter, r *http.Request) {
		req := URLRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

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
		json.NewEncoder(w).Encode(map[string]string{"shortened_url": resp.ShortUrl})
	})

	log.Fatal(http.ListenAndServe(":8080", nil)) // nil uses the default ServeMux
}
