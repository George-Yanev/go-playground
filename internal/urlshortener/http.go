package urlshortener

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func StartHttpServer(workCh chan<- WorkRequest) {
	http.HandleFunc("POST /short", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("got / request\n")

		req := URLRequest{}
		json.NewDecoder(r.Body).Decode(&req)

		doneCh := make(chan WorkResult, 1)

		work := WorkRequest{
			OriginalUrl: req.OriginalURL,
			DoneChan:    doneCh,
		}
		workCh <- work
		result := <-doneCh
		io.WriteString(w, result.ShortUrl)

	})

	log.Fatal(http.ListenAndServe(":8080", nil)) // nil uses the default ServeMux
}
