package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var ready int32 // 0 = not ready, 1 = ready

func main() {
	// Read configuration from environment
	region := os.Getenv("W")
	if region == "" {
		region = "unknown"
	}
	addr := envOr("ADDR", ":8080")

	log.Printf("starting server (region=%q) on %s\n", region, addr)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ready", readyHandler)
	mux.HandleFunc("/region", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"region": region})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"message": "hello from health-be", "region": region})
	})

	// Enable HTTP/2 over cleartext (h2c) so this binary can directly serve HTTP/2 if placed behind
	// a proxy that uses cleartext (useful for testing). For production TLS should be used.
	h2s := &http2.Server{}
	handler := h2c.NewHandler(loggingMiddleware(mux), h2s)

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
		// Good defaults
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// mark ready once server is listening
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}
	atomic.StoreInt32(&ready, 1)
	log.Println("server ready to accept connections")

	// Run server in background
	errCh := make(chan error, 1)
	go func() {
		if serr := server.Serve(ln); serr != nil && !errors.Is(serr, http.ErrServerClosed) {
			errCh <- serr
		}
		close(errCh)
	}()

	// Wait for signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigCh:
		log.Printf("signal received: %v; starting graceful shutdown", sig)
	case e := <-errCh:
		if e != nil {
			log.Fatalf("server error: %v", e)
		}
	}

	// stop accepting new requests
	atomic.StoreInt32(&ready, 0)

	// shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	} else {
		log.Println("server stopped gracefully")
	}
}

// healthHandler responds with a simple health payload.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// readyHandler reports whether the server is ready to accept traffic.
func readyHandler(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&ready) == 1 {
		writeJSON(w, http.StatusOK, map[string]string{"ready": "true"})
		return
	}
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{"ready": "false"})
}

// writeJSON is a convenience helper to return JSON responses with proper headers.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	// ensure stable output
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

// loggingMiddleware adds simple request logging including method, path and remote address.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rr := &responseRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rr, r)
		duration := time.Since(start)
		log.Printf("%s %s %d %s remote=%s user-agent=%q", r.Method, r.URL.Path, rr.status, duration, r.RemoteAddr, r.UserAgent())
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func envOr(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}
