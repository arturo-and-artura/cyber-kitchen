package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
	_ "modernc.org/sqlite"
)

// SQLite keeps a cached snapshot and commits each accepted state change in one SQL transaction.
type SQLite struct {
	mu    sync.RWMutex
	db    *sql.DB
	state domain.State
}

func NewSQLite(path string, initial domain.State) (*SQLite, error) {
	if path == "" {
		return nil, fmt.Errorf("database path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create database file: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("close database file: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return nil, fmt.Errorf("protect database file: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(1)
	store := &SQLite{db: db}
	if err := store.initialize(initial); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLite) Close() error {
	return s.db.Close()
}

func (s *SQLite) Snapshot() domain.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneState(s.state)
}

func (s *SQLite) Update(change func(*domain.State) error) (domain.State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := cloneState(s.state)
	if err := change(&next); err != nil {
		return domain.State{}, err
	}
	if err := s.persist(next); err != nil {
		return domain.State{}, err
	}
	s.state = next
	return cloneState(s.state), nil
}

func (s *SQLite) initialize(initial domain.State) error {
	if _, err := s.db.Exec(`
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;
CREATE TABLE IF NOT EXISTS app_state (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  selected_meal_id TEXT
);
CREATE TABLE IF NOT EXISTS household (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  name TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS household_members (
  position INTEGER PRIMARY KEY,
  id TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  initials TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS household_constraints (
  position INTEGER PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS household_goals (
  position INTEGER PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS inventory (
  position INTEGER NOT NULL,
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  amount REAL NOT NULL,
  unit TEXT NOT NULL,
  category TEXT NOT NULL,
  low_at REAL NOT NULL
);
CREATE TABLE IF NOT EXISTS meals (
  position INTEGER NOT NULL,
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  reason TEXT NOT NULL,
  emoji TEXT NOT NULL,
  accent TEXT NOT NULL,
  minutes INTEGER NOT NULL,
  difficulty TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS meal_tags (
  meal_id TEXT NOT NULL REFERENCES meals(id) ON DELETE CASCADE,
  position INTEGER NOT NULL,
  value TEXT NOT NULL,
  PRIMARY KEY (meal_id, position)
);
CREATE TABLE IF NOT EXISTS meal_ingredients (
  meal_id TEXT NOT NULL REFERENCES meals(id) ON DELETE CASCADE,
  position INTEGER NOT NULL,
  inventory_id TEXT NOT NULL,
  name TEXT NOT NULL,
  amount REAL NOT NULL,
  unit TEXT NOT NULL,
  optional INTEGER NOT NULL,
  PRIMARY KEY (meal_id, position)
);
CREATE TABLE IF NOT EXISTS meal_steps (
  meal_id TEXT NOT NULL REFERENCES meals(id) ON DELETE CASCADE,
  position INTEGER NOT NULL,
  value TEXT NOT NULL,
  PRIMARY KEY (meal_id, position)
);
CREATE TABLE IF NOT EXISTS history (
  position INTEGER NOT NULL,
  id TEXT PRIMARY KEY,
  meal_id TEXT NOT NULL,
  meal_name TEXT NOT NULL,
  emoji TEXT NOT NULL,
  cooked_at TEXT NOT NULL,
  rating TEXT NOT NULL,
  note TEXT NOT NULL
);`); err != nil {
		return fmt.Errorf("initialize database schema: %w", err)
	}

	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM app_state").Scan(&count); err != nil {
		return fmt.Errorf("inspect database seed state: %w", err)
	}
	if count == 0 {
		if err := s.persist(cloneState(initial)); err != nil {
			return fmt.Errorf("seed database: %w", err)
		}
	}
	state, err := loadState(s.db)
	if err != nil {
		return fmt.Errorf("load database state: %w", err)
	}
	s.state = state
	return nil
}

func (s *SQLite) persist(state domain.State) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin state transaction: %w", err)
	}
	defer tx.Rollback()
	if err := replaceState(tx, state); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit state transaction: %w", err)
	}
	return nil
}

func replaceState(tx *sql.Tx, state domain.State) error {
	for _, table := range []string{"meal_tags", "meal_ingredients", "meal_steps", "history", "meals", "inventory", "household_members", "household_constraints", "household_goals", "household", "app_state"} {
		if _, err := tx.Exec("DELETE FROM " + table); err != nil {
			return fmt.Errorf("clear %s: %w", table, err)
		}
	}

	var selected any
	if state.SelectedMealID != nil {
		selected = *state.SelectedMealID
	}
	if _, err := tx.Exec("INSERT INTO app_state (id, selected_meal_id) VALUES (1, ?)", selected); err != nil {
		return fmt.Errorf("write app state: %w", err)
	}
	if _, err := tx.Exec("INSERT INTO household (id, name) VALUES (1, ?)", state.Household.Name); err != nil {
		return fmt.Errorf("write household: %w", err)
	}
	for position, member := range state.Household.Members {
		if _, err := tx.Exec("INSERT INTO household_members (position, id, name, initials) VALUES (?, ?, ?, ?)", position, member.ID, member.Name, member.Initials); err != nil {
			return fmt.Errorf("write household member: %w", err)
		}
	}
	for position, value := range state.Household.Constraints {
		if _, err := tx.Exec("INSERT INTO household_constraints (position, value) VALUES (?, ?)", position, value); err != nil {
			return fmt.Errorf("write household constraint: %w", err)
		}
	}
	for position, value := range state.Household.Goals {
		if _, err := tx.Exec("INSERT INTO household_goals (position, value) VALUES (?, ?)", position, value); err != nil {
			return fmt.Errorf("write household goal: %w", err)
		}
	}
	for position, item := range state.Inventory {
		if _, err := tx.Exec("INSERT INTO inventory (position, id, name, amount, unit, category, low_at) VALUES (?, ?, ?, ?, ?, ?, ?)", position, item.ID, item.Name, item.Amount, item.Unit, item.Category, item.LowAt); err != nil {
			return fmt.Errorf("write inventory item: %w", err)
		}
	}
	for position, meal := range state.Meals {
		if _, err := tx.Exec("INSERT INTO meals (position, id, name, description, reason, emoji, accent, minutes, difficulty) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", position, meal.ID, meal.Name, meal.Description, meal.Reason, meal.Emoji, meal.Accent, meal.Minutes, meal.Difficulty); err != nil {
			return fmt.Errorf("write meal: %w", err)
		}
		for tagPosition, tag := range meal.Tags {
			if _, err := tx.Exec("INSERT INTO meal_tags (meal_id, position, value) VALUES (?, ?, ?)", meal.ID, tagPosition, tag); err != nil {
				return fmt.Errorf("write meal tag: %w", err)
			}
		}
		for ingredientPosition, ingredient := range meal.Ingredients {
			if _, err := tx.Exec("INSERT INTO meal_ingredients (meal_id, position, inventory_id, name, amount, unit, optional) VALUES (?, ?, ?, ?, ?, ?, ?)", meal.ID, ingredientPosition, ingredient.InventoryID, ingredient.Name, ingredient.Amount, ingredient.Unit, ingredient.Optional); err != nil {
				return fmt.Errorf("write meal ingredient: %w", err)
			}
		}
		for stepPosition, step := range meal.Steps {
			if _, err := tx.Exec("INSERT INTO meal_steps (meal_id, position, value) VALUES (?, ?, ?)", meal.ID, stepPosition, step); err != nil {
				return fmt.Errorf("write meal step: %w", err)
			}
		}
	}
	for position, entry := range state.History {
		if _, err := tx.Exec("INSERT INTO history (position, id, meal_id, meal_name, emoji, cooked_at, rating, note) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", position, entry.ID, entry.MealID, entry.MealName, entry.Emoji, entry.CookedAt.UTC().Format(time.RFC3339Nano), entry.Rating, entry.Note); err != nil {
			return fmt.Errorf("write history entry: %w", err)
		}
	}
	return nil
}

func loadState(db *sql.DB) (domain.State, error) {
	var state domain.State
	var selected sql.NullString
	if err := db.QueryRow("SELECT selected_meal_id FROM app_state WHERE id = 1").Scan(&selected); err != nil {
		return state, err
	}
	if selected.Valid {
		state.SelectedMealID = &selected.String
	}
	if err := db.QueryRow("SELECT name FROM household WHERE id = 1").Scan(&state.Household.Name); err != nil {
		return state, err
	}

	rows, err := db.Query("SELECT id, name, initials FROM household_members ORDER BY position")
	if err != nil {
		return state, err
	}
	for rows.Next() {
		var value domain.HouseholdMember
		if err := rows.Scan(&value.ID, &value.Name, &value.Initials); err != nil {
			rows.Close()
			return state, err
		}
		state.Household.Members = append(state.Household.Members, value)
	}
	if err := rows.Close(); err != nil {
		return state, err
	}
	if err := loadStrings(db, "SELECT value FROM household_constraints ORDER BY position", &state.Household.Constraints); err != nil {
		return state, err
	}
	if err := loadStrings(db, "SELECT value FROM household_goals ORDER BY position", &state.Household.Goals); err != nil {
		return state, err
	}

	rows, err = db.Query("SELECT id, name, amount, unit, category, low_at FROM inventory ORDER BY position")
	if err != nil {
		return state, err
	}
	for rows.Next() {
		var value domain.InventoryItem
		if err := rows.Scan(&value.ID, &value.Name, &value.Amount, &value.Unit, &value.Category, &value.LowAt); err != nil {
			rows.Close()
			return state, err
		}
		state.Inventory = append(state.Inventory, value)
	}
	if err := rows.Close(); err != nil {
		return state, err
	}

	rows, err = db.Query("SELECT id, name, description, reason, emoji, accent, minutes, difficulty FROM meals ORDER BY position")
	if err != nil {
		return state, err
	}
	for rows.Next() {
		var meal domain.Meal
		if err := rows.Scan(&meal.ID, &meal.Name, &meal.Description, &meal.Reason, &meal.Emoji, &meal.Accent, &meal.Minutes, &meal.Difficulty); err != nil {
			rows.Close()
			return state, err
		}
		state.Meals = append(state.Meals, meal)
	}
	if err := rows.Close(); err != nil {
		return state, err
	}
	for index := range state.Meals {
		meal := &state.Meals[index]
		if err := loadStrings(db, "SELECT value FROM meal_tags WHERE meal_id = ? ORDER BY position", &meal.Tags, meal.ID); err != nil {
			return state, err
		}
		ingredients, err := db.Query("SELECT inventory_id, name, amount, unit, optional FROM meal_ingredients WHERE meal_id = ? ORDER BY position", meal.ID)
		if err != nil {
			return state, err
		}
		for ingredients.Next() {
			var ingredient domain.MealIngredient
			if err := ingredients.Scan(&ingredient.InventoryID, &ingredient.Name, &ingredient.Amount, &ingredient.Unit, &ingredient.Optional); err != nil {
				ingredients.Close()
				return state, err
			}
			meal.Ingredients = append(meal.Ingredients, ingredient)
		}
		if err := ingredients.Close(); err != nil {
			return state, err
		}
		if err := loadStrings(db, "SELECT value FROM meal_steps WHERE meal_id = ? ORDER BY position", &meal.Steps, meal.ID); err != nil {
			return state, err
		}
	}

	rows, err = db.Query("SELECT id, meal_id, meal_name, emoji, cooked_at, rating, note FROM history ORDER BY position")
	if err != nil {
		return state, err
	}
	for rows.Next() {
		var entry domain.HistoryEntry
		var cookedAt string
		if err := rows.Scan(&entry.ID, &entry.MealID, &entry.MealName, &entry.Emoji, &cookedAt, &entry.Rating, &entry.Note); err != nil {
			rows.Close()
			return state, err
		}
		entry.CookedAt, err = time.Parse(time.RFC3339Nano, cookedAt)
		if err != nil {
			rows.Close()
			return state, fmt.Errorf("parse history time: %w", err)
		}
		state.History = append(state.History, entry)
	}
	if err := rows.Close(); err != nil {
		return state, err
	}
	return cloneState(state), nil
}

func loadStrings(db *sql.DB, query string, destination *[]string, args ...any) error {
	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return err
		}
		*destination = append(*destination, value)
	}
	return rows.Err()
}
