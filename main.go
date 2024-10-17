package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL, time.Since(begin))
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthcheck", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Server is running.")
	})

	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		fmt.Fprintln(w, "User ID:", id)
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      Log(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	log.Println("Server is running on port 8080")
	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}
}
