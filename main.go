package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	region := os.Getenv("W")
	log.Println("region: ", region)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "%v\n---\n", region)
		for name, headers := range req.Header {
			for _, h := range headers {
				fmt.Fprintf(w, "%v: %v\n", name, h)
			}
		}
	})

	log.Println("running server on :8080 ...")

	if err := http.ListenAndServe(":8080", nil); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen:%+s\n", err)
	}
}
