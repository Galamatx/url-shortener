package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
)

type Server struct {
	store   Store
	baseURL string
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /shorten", s.handleShorten)
	mux.HandleFunc("GET /stats/{code}", s.handleStats)
	mux.HandleFunc("GET /{code}", s.handleRedirect)
	return mux
}

func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if !isValidURL(req.URL) {
		http.Error(w, "url must be a valid http:// or https:// address", http.StatusBadRequest)
		return
	}

	code, err := s.store.CreateLink(req.URL)
	if err != nil {
		log.Println("create link:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"code":      code,
		"short_url": s.baseURL + "/" + code,
	})
}

func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	target, err := s.store.GetURLAndCountClick(code)
	if errors.Is(err, ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Println("redirect:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, target, http.StatusFound)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	target, clicks, err := s.store.GetStats(code)
	if errors.Is(err, ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Println("stats:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"code":   code,
		"url":    target,
		"clicks": clicks,
	})
}

func isValidURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
