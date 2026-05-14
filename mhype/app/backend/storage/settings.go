package storage

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

type Settings struct {
	DisabledCrawlers []string `json:"disabledCrawlers"`
}

type SettingsStore struct {
	path string
}

func NewSettingsStore(dataDir string) (*SettingsStore, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	return &SettingsStore{path: filepath.Join(dataDir, "settings.json")}, nil
}

func (s *SettingsStore) Load() (Settings, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return Settings{DisabledCrawlers: []string{}}, nil
	}
	if err != nil {
		return Settings{}, err
	}
	var out Settings
	if err := json.Unmarshal(data, &out); err != nil {
		return Settings{}, err
	}
	if out.DisabledCrawlers == nil {
		out.DisabledCrawlers = []string{}
	}
	return out, nil
}

func (s *SettingsStore) IsActive(crawlerID string) bool {
	cur, err := s.Load()
	if err != nil {
		return true
	}
	return !slices.Contains(cur.DisabledCrawlers, crawlerID)
}

func (s *SettingsStore) SetActive(crawlerID string, active bool) error {
	cur, err := s.Load()
	if err != nil {
		return err
	}
	has := slices.Contains(cur.DisabledCrawlers, crawlerID)
	switch {
	case active && has:
		cur.DisabledCrawlers = slices.DeleteFunc(cur.DisabledCrawlers, func(id string) bool { return id == crawlerID })
	case !active && !has:
		cur.DisabledCrawlers = append(cur.DisabledCrawlers, crawlerID)
	default:
		return nil
	}
	return s.write(cur)
}

func (s *SettingsStore) write(cfg Settings) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, s.path); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}
