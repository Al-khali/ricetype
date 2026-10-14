package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Record struct {
	CategoryKey string    `json:"category_key"`
	ModeLabel   string    `json:"mode_label"`
	Language    string    `json:"language"`
	NetWPM      float64   `json:"net_wpm"`
	RawWPM      float64   `json:"raw_wpm"`
	Accuracy    float64   `json:"accuracy"`
	Consistency float64   `json:"consistency"`
	Rank        string    `json:"rank"`
	Timestamp   time.Time `json:"timestamp"`
}

type Store struct {
	ThemeIndex   int               `json:"theme_index"`
	TotalTests   int               `json:"total_tests"`
	TotalSeconds float64           `json:"total_seconds"`
	PersonalBest map[string]Record `json:"personal_best"`
	History      []Record          `json:"history"`
	filePath     string
}

func configFilePath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir = "."
	}
	return filepath.Join(dir, "ricetype", "stats.json")
}

func Load() *Store {
	path := configFilePath()
	s := &Store{
		PersonalBest: make(map[string]Record),
		History:      make([]Record, 0, 16),
		filePath:     path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	_ = json.Unmarshal(data, s)
	if s.PersonalBest == nil {
		s.PersonalBest = make(map[string]Record)
	}
	s.filePath = path
	return s
}

func (s *Store) Save() error {
	if s.filePath == "" {
		s.filePath = configFilePath()
	}
	if err := os.MkdirAll(filepath.Dir(s.filePath), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0o644)
}

// RecordSession saves a finished typing test and returns true if it beat the category Personal Best.
func (s *Store) RecordSession(rec Record, durationSec float64, suddenDeathFailed bool) bool {
	s.TotalTests++
	s.TotalSeconds += durationSec

	// Prepend to recent history (keep max 12 entries)
	s.History = append([]Record{rec}, s.History...)
	if len(s.History) > 12 {
		s.History = s.History[:12]
	}

	isNewPB := false
	if !suddenDeathFailed && rec.NetWPM > 0 {
		prev, exists := s.PersonalBest[rec.CategoryKey]
		if !exists || rec.NetWPM > prev.NetWPM {
			s.PersonalBest[rec.CategoryKey] = rec
			isNewPB = true
		}
	}

	_ = s.Save()
	return isNewPB
}

func (s *Store) AverageWPM() float64 {
	if len(s.History) == 0 {
		return 0
	}
	sum := 0.0
	for _, r := range s.History {
		sum += r.NetWPM
	}
	return sum / float64(len(s.History))
}

func (s *Store) BestOverallWPM() float64 {
	best := 0.0
	for _, r := range s.PersonalBest {
		if r.NetWPM > best {
			best = r.NetWPM
		}
	}
	return best
}
