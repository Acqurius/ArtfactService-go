package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	dbPath := "files.db"
	
	// Check if database already exists
	if _, err := os.Stat(dbPath); err == nil {
		fmt.Printf("Database '%s' already exists. Do you want to recreate it? (y/N): ", dbPath)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Initialization cancelled.")
			return
		}
		// Remove existing database
		if err := os.Remove(dbPath); err != nil {
			log.Fatalf("Failed to remove existing database: %v", err)
		}
		fmt.Println("Removed existing database.")
	}

	// Create new database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create Artifacts table
	artifactsTable := `
	CREATE TABLE IF NOT EXISTS Artifacts (
		uuid TEXT PRIMARY KEY,
		filename TEXT NOT NULL,
		content_type TEXT NOT NULL,
		size BIGINT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(artifactsTable); err != nil {
		log.Fatalf("Failed to create Artifacts table: %v", err)
	}
	fmt.Println("✓ Created table: Artifacts")

	// Create tokens table
	tokensTable := `
	CREATE TABLE IF NOT EXISTS tokens (
		token TEXT PRIMARY KEY,
		artifact_uuid TEXT NOT NULL,
		valid_from TIMESTAMP,
		valid_to TIMESTAMP,
		max_downloads BIGINT,
		current_downloads BIGINT DEFAULT 0,
		allowed_cidr TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(artifact_uuid) REFERENCES Artifacts(uuid)
	);`

	if _, err := db.Exec(tokensTable); err != nil {
		log.Fatalf("Failed to create tokens table: %v", err)
	}
	fmt.Println("✓ Created table: tokens")

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_artifacts_created_at ON Artifacts(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_tokens_artifact_uuid ON tokens(artifact_uuid);",
		"CREATE INDEX IF NOT EXISTS idx_tokens_valid_to ON tokens(valid_to);",
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			log.Fatalf("Failed to create index: %v", err)
		}
	}
	fmt.Println("✓ Created indexes")

	fmt.Printf("\n✅ Database '%s' initialized successfully!\n", dbPath)
}
