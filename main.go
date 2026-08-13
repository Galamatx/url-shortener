package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	port := getEnv("PORT", "8080")
	baseURL := getEnv("BASE_URL", "http://localhost:"+port)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("cannot connect to database: ", err)
	}

	srv := &Server{store: NewPostgresStore(db), baseURL: baseURL}

	log.Println("listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, srv.routes()))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
