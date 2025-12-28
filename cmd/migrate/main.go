package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Fallback for dev
		dbURL = "postgres://postgres:postgres@localhost:5432/wordle?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	files, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fmt.Printf("Applying migration %s...\n", file)
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read file %s: %v", file, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			fmt.Printf("Migration %s failed (might be already applied): %v\n", file, err)
		} else {
			fmt.Printf("Migration %s applied successfully\n", file)
		}
	}
}
