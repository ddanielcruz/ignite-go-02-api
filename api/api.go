package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/exp/rand"
)

func NewHandler(db map[string]string) http.Handler {
	r := chi.NewMux()

	// Middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	// Routes
	r.Post("/shorten", handleShortenLink(db))
	r.Get("/{code}", handleGetLink(db))

	return r
}

type ShortenLinkRequest struct {
	URL string `json:"url"`
}

type Response struct {
	Error string `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}

func handleShortenLink(db map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body ShortenLinkRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			sendJSON(w, Response{Error: "invalid request body"}, http.StatusUnprocessableEntity)
			return
		}

		parsedURL, ok := isURL(body.URL)
		if !ok {
			sendJSON(w, Response{Error: "invalid URL"}, http.StatusBadRequest)
			return
		}

		code := generateCode()
		db[code] = parsedURL.String()

		sendJSON(w, Response{Data: code}, http.StatusCreated)
	}
}

func handleGetLink(db map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")
		url, ok := db[code]

		if !ok {
			// Return a not found response instead of JSON because the code is not a valid URL, and the
			// person using the service doesn't need to know that. In practice, the user should be
			// redirected to the home page.
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, url, http.StatusPermanentRedirect)
	}
}

func sendJSON(w http.ResponseWriter, resp Response, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func isURL(str string) (*url.URL, bool) {
	url, err := url.Parse(str)
	return url, err == nil && url.Scheme != "" && url.Host != ""
}

const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateCode() string {
	const length = 8
	bytes := make([]byte, length)

	for i := range bytes {
		bytes[i] = characters[rand.Intn(len(characters))]
	}

	return string(bytes)
}
