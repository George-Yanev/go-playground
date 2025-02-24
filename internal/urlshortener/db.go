package urlshortener

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

//go:embed init.sql
var initSQL string

type UrlMapping struct {
	db *sql.DB
}

type SeedsDb struct {
	db *sql.DB
}

func NewUrlMapping(db *sql.DB) *UrlMapping {
	return &UrlMapping{db: db}
}

func (u *UrlMapping) Create(orig_url, short_url, seed string, counter int) error {
	_, err := u.db.Exec(
		"INSERT INTO url_mapping (original_url, short_url, seed, counter, created_at) "+
			"VALUES (?,?,?,?,?)",
		orig_url, short_url, seed, counter, time.Now().UTC(),
	)
	return err
}

func NewSeedsDb(db *sql.DB) *SeedsDb {
	return &SeedsDb{db: db}
}

func (s *SeedsDb) Create(seed string) error {
	_, err := s.db.Exec("INSERT INTO seeds (seed) VALUES (?)", seed)
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

	r, err := db.Query("SELECT COUNT(*) FROM seeds")
	if err != nil {
		return nil, fmt.Errorf("Cannot get the seeds table count. Error: %w", err)
	}
	defer r.Close()

	var count int
	if r.Next() {
		err = r.Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("Failed to scan count: %w", err)
		}
	} else {
		return nil, fmt.Errorf("No count returned from query")
	}

	if err := r.Err(); err != nil {
		return nil, fmt.Errorf("Error during result set iteration: %w", err)
	}

	if count == 0 {
		s := NewSeedsDb(db)
		seeds := generateSeeds()
		// log.Printf("generateSeeds output: %v. Len: %d\n", seeds, len(seeds))
		for _, seed := range seeds {
			err := s.Create(seed.Seed)
			if err != nil {
				return nil, fmt.Errorf("Cannot create seed: %s. Error: %w", seed.Seed, err)
			}
		}
	}

	log.Println("Database initialized successfully")
	return db, nil
}

func generateSeeds() Seeds {
	// generate a small amount of seeds
	seed_letters := []string{"a", "b", "c"}
	seeds := make(Seeds, 0)
	// TODO n^3 looks bad. Can I improve it?
	for i := 0; i < len(seed_letters); i++ {
		for j := 0; j < len(seed_letters); j++ {
			for k := 0; k < len(seed_letters); k++ {
				seeds = append(seeds, Seed{
					Seed:    fmt.Sprintf("%s%s%s", seed_letters[i], seed_letters[j], seed_letters[k]),
					Counter: 0,
				})
			}
		}
	}
	return seeds
}
