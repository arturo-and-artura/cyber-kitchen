package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Postgres stores the single-household pilot as one validated aggregate. Domain
// changes replace that aggregate atomically in one transaction.
type Postgres struct {
	mu    sync.RWMutex
	db    *sql.DB
	state domain.State
}

func NewPostgres(databaseURL string, initial domain.State) (*Postgres, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	store := &Postgres{db: db}
	if err := store.initialize(initial); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Postgres) Close() error { return s.db.Close() }

func (s *Postgres) Snapshot() domain.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneState(s.state)
}

func (s *Postgres) Update(change func(*domain.State) error) (domain.State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := cloneState(s.state)
	if err := change(&next); err != nil {
		return domain.State{}, err
	}
	encoded, err := json.Marshal(next)
	if err != nil {
		return domain.State{}, fmt.Errorf("encode state: %w", err)
	}
	tx, err := s.db.Begin()
	if err != nil {
		return domain.State{}, fmt.Errorf("begin state transaction: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE kitchen_state SET state = $1::jsonb WHERE singleton = TRUE`, encoded); err != nil {
		return domain.State{}, fmt.Errorf("write state: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.State{}, fmt.Errorf("commit state transaction: %w", err)
	}
	s.state = next
	return cloneState(next), nil
}

func (s *Postgres) initialize(initial domain.State) error {
	if err := s.db.Ping(); err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	if _, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS kitchen_state (
  singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
  state JSONB NOT NULL
)`); err != nil {
		return fmt.Errorf("initialize database schema: %w", err)
	}
	encoded, err := json.Marshal(initial)
	if err != nil {
		return fmt.Errorf("encode seed state: %w", err)
	}
	if _, err := s.db.Exec(`INSERT INTO kitchen_state (singleton, state) VALUES (TRUE, $1::jsonb) ON CONFLICT (singleton) DO NOTHING`, encoded); err != nil {
		return fmt.Errorf("seed database: %w", err)
	}
	var persisted []byte
	if err := s.db.QueryRow(`SELECT state FROM kitchen_state WHERE singleton = TRUE`).Scan(&persisted); err != nil {
		return fmt.Errorf("load database state: %w", err)
	}
	if err := json.Unmarshal(persisted, &s.state); err != nil {
		return fmt.Errorf("decode database state: %w", err)
	}
	s.state = cloneState(s.state)
	return nil
}
