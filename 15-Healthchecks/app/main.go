package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	ready = false
	mutex = sync.Mutex{}
)

func main() {
	// Simulate application startup delay
	go func() {
		time.Sleep(10 * time.Second) // Simulate startup time
		mutex.Lock()
		ready = true
		mutex.Unlock()
		log.Println("Application is ready!")
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, Kubernetes!")
	})

	// Liveness probe - Always returns 200 if the server is running
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// Readiness probe - Returns 500 until app is ready
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		if ready {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Ready")
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, "Not Ready")
		}
	})

	// Startup probe - Returns 200 only after startup delay
	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		if ready {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Started")
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, "Starting")
		}
	})

	log.Println("Server is starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
