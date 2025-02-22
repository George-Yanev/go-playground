package urlshortener

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
)

//go:embed init.sql
var initSQL string

type UrlMapping struct {
	db *sql.DB
}

func NewUrlMapping(db *sql.DB) *UrlMapping {
	return &UrlMapping{db: db}
}

func (u *UrlMapping) Create(orig_url, short_url, seed string, counter int) error {
	_, err := u.db.Exec(
		"INSERT INTO url_mapping (orig_url, short_url, seed, counter) "+
			"VALUES (?,?,?,?)",
		orig_url, short_url, seed, counter,
	)
	return err
}

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:urlshortener.db")
	if err != nil {
		return nil, fmt.Errorf("Unable to create/open the database: %w", err)
	}

	_, err = db.Exec(initSQL)
	if err != nil {
		return nil, fmt.Errorf("Failed to execute init.sql: %w", err)
	}

	log.Println("Database initialized successfully")
	return db, nil
}
