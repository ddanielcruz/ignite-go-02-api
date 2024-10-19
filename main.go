package main

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type User struct {
	ID       uint64 `json:"id,string"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Password string `json:"-"`
}

type Response struct {
	Error string `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}

func sendJSON(w http.ResponseWriter, response Response, status int) {
	data, err := json.Marshal(response)
	if err != nil {
		slog.Error("Error marshalling response:", "error", err)
		sendJSON(w, Response{Error: "Internal server error."}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		slog.Error("Error writing response:", "error", err)
	}
}

func main() {
	// Set default logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Create new router
	r := chi.NewMux()

	// Middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	// Memory store
	db := map[uint64]User{
		1: {
			ID:       1,
			Name:     "Daniel",
			Role:     "admin",
			Password: "123",
		},
		2: {
			ID:       2,
			Name:     "Camila",
			Role:     "user",
			Password: "456",
		},
	}

	// Routes
	r.Group(func(r chi.Router) {
		r.Use(jsonMiddleware)
		r.Get("/users/{id:[1-9][0-9]*}", handleGetUser(db))
		r.Post("/users", handleCreateUser(db))
	})

	// Start server
	slog.Info(
		"Starting server on port 8080",
		slog.String("version", "1.0.0"),
		slog.String("env", "dev"),
		slog.String("port", "8080"),
	)
	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func handleGetUser(db map[uint64]User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, _ := strconv.ParseUint(idStr, 10, 64)

		user, ok := db[id]
		if !ok {
			sendJSON(w, Response{Error: "User not found."}, http.StatusNotFound)
			return
		}

		sendJSON(w, Response{Data: user}, http.StatusOK)
	}
}

func handleCreateUser(db map[uint64]User) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024) // 1MB max
		data, err := io.ReadAll(r.Body)

		if err != nil {
			var maxErr *http.MaxBytesError

			if errors.As(err, &maxErr) {
				sendJSON(w, Response{Error: "Request body too large."}, http.StatusRequestEntityTooLarge)
			} else {
				slog.Error("Error reading request body:", "error", err)
				sendJSON(w, Response{Error: "Internal server error."}, http.StatusInternalServerError)
			}
			return
		}

		var user User
		if err := json.Unmarshal(data, &user); err != nil {
			slog.Error("Error unmarshalling user:", "error", err)
			sendJSON(w, Response{Error: "Invalid request payload."}, http.StatusUnprocessableEntity)
			return
		}

		user.ID = uint64(len(db) + 1)
		user.Role = "user"
		db[user.ID] = user

		sendJSON(w, Response{Data: user}, http.StatusCreated)
	}
}
