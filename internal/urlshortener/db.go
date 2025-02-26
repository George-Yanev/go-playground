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

func (u *UrlMapping) GetSeedCounter(seed string) (int, error) {
	var counter int
	err := u.db.QueryRow("Select MAX(counter) FROM url_mapping WHERE seed = ?", seed).Scan(&counter)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // counter will start from 0
		}
		return -1, fmt.Errorf("unable to get seed counter from url_mapping: %w", err)
	}

	return counter, nil
}

func NewSeedsDb(db *sql.DB) *SeedsDb {
	return &SeedsDb{db: db}
}

func (s *SeedsDb) Create(seed string) error {
	_, err := s.db.Exec("INSERT INTO seeds (seed) VALUES (?)", seed)
	return err
}

func (s *SeedsDb) SetSeedStatusAndCounter(seed string, counter, status int) error {
	_, err := s.db.Exec("UPDATE seeds SET counter_used = ?, status = ? WHERE seed = ?", counter, status, seed)
	if err != nil {
		return fmt.Errorf("updating seed %s: %w", seed, err)
	}
	return nil
}

func (s *SeedsDb) SelectSeedByStatus(status int) (Seeds, error) {
	var seeds Seeds
	r, err := s.db.Query("SELECT seed, counter_used, counter_size FROM seeds WHERE status = ?", status)
	if err != nil {
		return nil, fmt.Errorf("Selecting seeds by status: %d. Err: %w", status, err)
	}
	defer r.Close()

	for r.Next() {
		var s Seed
		if err := r.Scan(&s.Seed, &s.CounterUsed, &s.CounterSize); err != nil {
			return nil, fmt.Errorf("Scanning seed row: %w", err)
		}
		seeds = append(seeds, s)
	}
	if err := r.Err(); err != nil {
		return nil, fmt.Errorf("Error iterating seed rows: %w", err)
	}

	return seeds, nil
}

// func (s *SeedsDb) QueryWorkerUsedSeeds(worker string) (Seeds, error) {
// 	var seed Seed
// 	r, err := s.db.Query("SELECT * from seeds WHERE lease_holder = ? AND status = 1", client)
// 	if err != nil {
// 		return Seeds{}, fmt.Errorf("Error quering for worker used seeds: %w", err)
// 	}

// 	if r.Scan(&seed)

// }

// func (s *SeSeedsDb) ResetSeedStatusAndSetLastCounter(seed string) error {

// }

func (s *SeedsDb) Acquire(holder string) (Seed, error) {
	var acquiredSeed Seed

	query := `
UPDATE seeds
SET
    status = 1,
    lease_holder = ?,
    lease_taken = datetime('now')
WHERE
    status = 0
    AND rowid = (
        SELECT rowid
        FROM seeds
        WHERE status = 0
        ORDER BY lease_taken ASC NULLS LAST
        LIMIT 1
    )
RETURNING seed, counter_used, counter_size
`
	err := s.db.QueryRow(query, holder).Scan(&acquiredSeed.Seed, &acquiredSeed.CounterUsed, &acquiredSeed.CounterSize)
	if err != nil {
		if err == sql.ErrNoRows {
			return Seed{}, fmt.Errorf("No seeds available for acquisition")
		}
		return Seed{}, fmt.Errorf("Failed to acquire seed: %w", err)
	}

	return acquiredSeed, nil
}

// func (s *SeedsDb) SelectSeedsByLeaseHolder(holder string) (Seeds, error) {
// 	var seeds Seeds
// 	r, err := s.db.Query("SELECT seed, counter FROM seeds WHERE lease_holder = ? ORDER BY counter DESC LIMIT 1", holder)
// 	if err != nil {
// 		return nil, fmt.Errorf("Selecting seeds by lease_holder: %w", err)
// 	}
// 	defer r.Close()

// 	for r.Next() {
// 		var s Seed
// 		if err := r.Scan(&s.Seed, &s.Counter); err != nil {
// 			return nil, fmt.Errorf("Scanning seed row: %w", err)
// 		}
// 		seeds = append(seeds, s)
// 	}
// 	if err := r.Err(); err != nil {
// 		return nil, fmt.Errorf("Error iterating seed rows: %w", err)
// 	}
// 	return seeds, nil
// }

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
					Seed:        fmt.Sprintf("%s%s%s", seed_letters[i], seed_letters[j], seed_letters[k]),
					CounterUsed: 0,
				})
			}
		}
	}
	return seeds
}
