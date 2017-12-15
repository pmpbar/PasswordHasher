package main

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	serverAddr = ":8080"
)

func main() {
	// Place code in seperate function for testing
	server()
}

func server() {
	// System interrupt signal
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	// HTTP server on port 8080
	srv := &http.Server{Addr: serverAddr}

	// Root Handler function for hashing passwords
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get POST data from request
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		// Get password data from post
		pass := r.Form.Get("password")
		h := sha512.New()
		// Write plaintext password to SHA512 hash
		_, err = h.Write([]byte(pass))
		if err != nil {
			log.Println(err)
		}
		// Wait 5 seconds before responding
		time.Sleep(5 * time.Second)
		// Write back base64 encoded hash
		_, err = w.Write([]byte(base64.StdEncoding.EncodeToString(h.Sum(nil))))
		if err != nil {
			log.Println(err)
		}
	})

	http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		quit <- os.Interrupt
		_, err := w.Write([]byte("Performing graceful shutdown"))
		if err != nil {
			log.Println(err)
		}
	})
	// Channel to control main thread on shutdown
	done := make(chan bool)

	// Graceful shutdown thread
	go func() {
		// Recieved SIGINT
		<-quit

		// Add 5s deadline to context to finish requests on shutdown
		d := time.Now().Add(5 * time.Second)
		ctx, cancel := context.WithDeadline(context.Background(), d)

		// Fallback
		defer cancel()

		log.Println("Shutting down server...")
		// Graceful shutdown of HTTP server
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("could not shutdown: %v", err)
		}
		// Send message to main thread that it can exit
		close(done)
	}()

	// Start server
	log.Println("Server started at", serverAddr)
	err := srv.ListenAndServe()

	if err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}

	log.Println("Server gracefully stopped")
	<-done
}
