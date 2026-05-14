package orchestrator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type RunState struct {
	LastAttemptAt int64  `json:"lastAttemptAt"`
	LastSuccessAt int64  `json:"lastSuccessAt"`
	LastFailureAt int64  `json:"lastFailureAt"`
	LastError     string `json:"lastError,omitempty"`
}

type stateFile struct {
	Crawlers map[string]RunState `json:"crawlers"`
}

type StateStore struct {
	path string
	mu   sync.Mutex
	data map[string]RunState
}

func NewStateStore(dataDir string) (*StateStore, error) {
	path := filepath.Join(dataDir, "orchestrator.state.json")
	s := &StateStore{path: path, data: make(map[string]RunState)}
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return s, nil
		}
		return nil, fmt.Errorf("read orchestrator state: %w", err)
	}
	var f stateFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("parse orchestrator state: %w", err)
	}
	if f.Crawlers != nil {
		s.data = f.Crawlers
	}
	return s, nil
}

func (s *StateStore) Get(crawlerID string) RunState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data[crawlerID]
}

func (s *StateStore) RecordAttempt(crawlerID string, atMs int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rs := s.data[crawlerID]
	rs.LastAttemptAt = atMs
	s.data[crawlerID] = rs
	return s.flushLocked()
}

func (s *StateStore) RecordSuccess(crawlerID string, atMs int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rs := s.data[crawlerID]
	rs.LastAttemptAt = atMs
	rs.LastSuccessAt = atMs
	s.data[crawlerID] = rs
	return s.flushLocked()
}

func (s *StateStore) RecordFailure(crawlerID string, atMs int64, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rs := s.data[crawlerID]
	rs.LastAttemptAt = atMs
	rs.LastFailureAt = atMs
	rs.LastError = errMsg
	s.data[crawlerID] = rs
	return s.flushLocked()
}

func LoadStateFromDisk(dataDir string) (map[string]RunState, error) {
	path := filepath.Join(dataDir, "orchestrator.state.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read orchestrator state: %w", err)
	}
	var f stateFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("parse orchestrator state: %w", err)
	}
	return f.Crawlers, nil
}

func (s *StateStore) flushLocked() error {
	f := stateFile{Crawlers: s.data}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
