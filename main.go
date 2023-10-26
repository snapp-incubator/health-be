package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	region := os.Getenv("W")
	log.Println("region: ", region)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "%v\n---\n", region)
		for name, headers := range req.Header {
			for _, h := range headers {
				fmt.Fprintf(w, "%v: %v\n", name, h)
			}
		}
	})

	h2s := &http2.Server{}
	h1s := &http.Server{
		Addr:    ":8080",
		Handler: h2c.NewHandler(mux, h2s),
	}

	log.Println("running server on :8080 ...")

	if err := h1s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen:%+s\n", err)
	}
}
