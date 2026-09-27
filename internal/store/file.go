package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
)

// File keeps the active state in memory and atomically persists every successful update.
type File struct {
	mu    sync.RWMutex
	path  string
	state domain.State
}

func NewFile(path string, initial domain.State) (*File, error) {
	if path == "" {
		return nil, fmt.Errorf("data file path is required")
	}

	file := &File{path: path}
	contents, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := json.Unmarshal(contents, &file.state); err != nil {
			return nil, fmt.Errorf("decode data file %q: %w", path, err)
		}
	case os.IsNotExist(err):
		file.state = cloneState(initial)
		if err := writeStateFile(path, file.state); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("read data file %q: %w", path, err)
	}

	return file, nil
}

func (f *File) Snapshot() domain.State {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return cloneState(f.state)
}

func (f *File) Update(change func(*domain.State) error) (domain.State, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	next := cloneState(f.state)
	if err := change(&next); err != nil {
		return domain.State{}, err
	}
	if err := writeStateFile(f.path, next); err != nil {
		return domain.State{}, err
	}

	f.state = next
	return cloneState(f.state), nil
}

func writeStateFile(path string, state domain.State) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return fmt.Errorf("create data directory %q: %w", directory, err)
	}

	temporary, err := os.CreateTemp(directory, ".cyber-kitchen-*.json")
	if err != nil {
		return fmt.Errorf("create temporary data file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return fmt.Errorf("protect temporary data file: %w", err)
	}

	encoder := json.NewEncoder(temporary)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(state); err != nil {
		temporary.Close()
		return fmt.Errorf("encode data file: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync data file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close data file: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace data file %q: %w", path, err)
	}

	return nil
}
