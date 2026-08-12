package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeStore struct {
	urls   map[string]string
	clicks map[string]int
}

func newFakeStore() *fakeStore {
	return &fakeStore{urls: map[string]string{}, clicks: map[string]int{}}
}

func (f *fakeStore) CreateLink(url string) (string, error) {
	code := fmt.Sprintf("code%d", len(f.urls)+1)
	f.urls[code] = url
	return code, nil
}

func (f *fakeStore) GetURLAndCountClick(code string) (string, error) {
	url, ok := f.urls[code]
	if !ok {
		return "", ErrNotFound
	}
	f.clicks[code]++
	return url, nil
}

func (f *fakeStore) GetStats(code string) (string, int, error) {
	url, ok := f.urls[code]
	if !ok {
		return "", 0, ErrNotFound
	}
	return url, f.clicks[code], nil
}

func newTestServer() (*Server, *fakeStore) {
	store := newFakeStore()
	return &Server{store: store, baseURL: "http://short.test"}, store
}

func TestShorten(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"valid https url", `{"url":"https://example.com/page"}`, http.StatusCreated},
		{"valid http url", `{"url":"http://example.com"}`, http.StatusCreated},
		{"broken json", `{"url":`, http.StatusBadRequest},
		{"empty url", `{"url":""}`, http.StatusBadRequest},
		{"no scheme", `{"url":"example.com"}`, http.StatusBadRequest},
		{"ftp scheme", `{"url":"ftp://example.com"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := newTestServer()
			req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			srv.routes().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestShortenReturnsShortURL(t *testing.T) {
	srv, _ := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(`{"url":"https://example.com"}`))
	rec := httptest.NewRecorder()

	srv.routes().ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), `"short_url":"http://short.test/code1"`) {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestRedirect(t *testing.T) {
	srv, store := newTestServer()
	store.urls["abc"] = "https://example.com/target"

	req := httptest.NewRequest(http.MethodGet, "/abc", nil)
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "https://example.com/target" {
		t.Errorf("Location = %q", got)
	}
	if store.clicks["abc"] != 1 {
		t.Errorf("clicks = %d, want 1", store.clicks["abc"])
	}
}

func TestRedirectNotFound(t *testing.T) {
	srv, _ := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	srv.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestStats(t *testing.T) {
	srv, store := newTestServer()
	store.urls["abc"] = "https://example.com"
	store.clicks["abc"] = 3

	req := httptest.NewRequest(http.MethodGet, "/stats/abc", nil)
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"clicks":3`) {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestStatsNotFound(t *testing.T) {
	srv, _ := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/stats/missing", nil)
	rec := httptest.NewRecorder()

	srv.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestIsValidURL(t *testing.T) {
	valid := []string{"http://a.com", "https://a.com/x?y=1"}
	invalid := []string{"", "a.com", "ftp://a.com", "https://", "javascript:alert(1)"}

	for _, u := range valid {
		if !isValidURL(u) {
			t.Errorf("%q should be valid", u)
		}
	}
	for _, u := range invalid {
		if isValidURL(u) {
			t.Errorf("%q should be invalid", u)
		}
	}
}
