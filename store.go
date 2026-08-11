package main

import (
	"database/sql"
	"errors"
	"math/rand/v2"

	"github.com/lib/pq"
)

var ErrNotFound = errors.New("link not found")

type Store interface {
	CreateLink(url string) (string, error)
	GetURLAndCountClick(code string) (string, error)
	GetStats(code string) (url string, clicks int, err error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomCode(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.IntN(len(letters))]
	}
	return string(b)
}

func (s *PostgresStore) CreateLink(url string) (string, error) {
	for attempt := 0; attempt < 5; attempt++ {
		code := randomCode(6)
		_, err := s.db.Exec(`INSERT INTO links (code, url) VALUES ($1, $2)`, code, url)
		if err == nil {
			return code, nil
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			continue
		}
		return "", err
	}
	return "", errors.New("could not generate a unique code")
}

func (s *PostgresStore) GetURLAndCountClick(code string) (string, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var id int
	var url string
	err = tx.QueryRow(`SELECT id, url FROM links WHERE code = $1`, code).Scan(&id, &url)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}

	if _, err := tx.Exec(`INSERT INTO clicks (link_id) VALUES ($1)`, id); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return url, nil
}

func (s *PostgresStore) GetStats(code string) (string, int, error) {
	var url string
	var clicks int
	err := s.db.QueryRow(`
		SELECT l.url, COUNT(c.id)
		FROM links l
		LEFT JOIN clicks c ON c.link_id = l.id
		WHERE l.code = $1
		GROUP BY l.id`, code).Scan(&url, &clicks)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, ErrNotFound
	}
	if err != nil {
		return "", 0, err
	}
	return url, clicks, nil
}
