package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	var err error
	// Switch to SQLite, simple file based DB
	connStr := "files.db"
	
	DB, err = sql.Open("sqlite", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database: ", err)
	}

	fmt.Println("Connected to SQLite database")

	createTable()
}

func createTable() {
	// SQLite syntax
	query := `
	CREATE TABLE IF NOT EXISTS Artifacts (
		uuid TEXT PRIMARY KEY,
		filename TEXT NOT NULL,
		content_type TEXT NOT NULL,
		size BIGINT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal("Failed to create table Artifacts: ", err)
	}
	fmt.Println("Table 'Artifacts' ensured")

	// Create tokens table
	queryTokens := `
	CREATE TABLE IF NOT EXISTS tokens (
		token TEXT PRIMARY KEY,
		artifact_uuid TEXT,
		valid_from TIMESTAMP,
		valid_to TIMESTAMP,
		max_downloads BIGINT,
		current_downloads BIGINT DEFAULT 0,
		allowed_cidr TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(artifact_uuid) REFERENCES Artifacts(uuid)
	);`

	_, err = DB.Exec(queryTokens)
	if err != nil {
		log.Fatal("Failed to create table tokens: ", err)
	}
	fmt.Println("Table 'tokens' ensured")
}
