package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	region := os.Getenv("W")
	if region == "" {
		region = "unknown"
	}
	log.Printf("Region: %s\n", region)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		var b strings.Builder

		// Write region and separator
		b.WriteString(region)
		b.WriteString("\n---\n")

		// Append all headers
		for name, values := range req.Header {
			for _, value := range values {
				b.WriteString(name)
				b.WriteString(": ")
				b.WriteString(value)
				b.WriteByte('\n')
			}
		}

		// Write to the response in one go
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(b.String()))
	})

	h2s := &http2.Server{}
	h1s := &http.Server{
		Addr:    ":8080",
		Handler: h2c.NewHandler(mux, h2s),
	}

	log.Println("Server running on :8080 ...")

	if err := h1s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %+v", err)
	}
}
