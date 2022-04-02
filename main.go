package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	region := os.Getenv("W")
	fmt.Println("region: ", region)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "%v\n---\n", region)
		for name, headers := range req.Header {
			for _, h := range headers {
				fmt.Fprintf(w, "%v: %v\n", name, h)
			}
		}
	})

	fmt.Println("running server on :8080 ...")
	http.ListenAndServe(":8080", nil)
}
