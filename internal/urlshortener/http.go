package urlshortener

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func StartHttpServer(workCh chan<- WorkRequest, shortUrlHost string) {
	http.HandleFunc("POST /short", func(w http.ResponseWriter, r *http.Request) {
		req := URLRequest{}
		json.NewDecoder(r.Body).Decode(&req)

		doneCh := make(chan WorkResult, 1)

		work := WorkRequest{
			OriginalUrl:  req.OriginalURL,
			ShortUrlHost: shortUrlHost,
			DoneCh:       doneCh,
		}
		workCh <- work
		result := <-doneCh
		io.WriteString(w, result.ShortUrl)

	})

	log.Fatal(http.ListenAndServe(":8080", nil)) // nil uses the default ServeMux
}
