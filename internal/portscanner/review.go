package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type ReviewStatus string

const (
	ReviewPending  ReviewStatus = "pending"
	ReviewApproved ReviewStatus = "approved"
	ReviewRejected ReviewStatus = "rejected"
)

type ReviewEntry struct {
	Port      int          `json:"port"`
	Proto     string       `json:"proto"`
	Status    ReviewStatus `json:"status"`
	Reviewer  string       `json:"reviewer"`
	Note      string       `json:"note,omitempty"`
	UpdatedAt time.Time    `json:"updated_at"`
}

func reviewKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type ReviewStore struct {
	dir     string
	entries map[string]ReviewEntry
}

func NewReviewStore(dir string) (*ReviewStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("review store: mkdir: %w", err)
	}
	s := &ReviewStore{dir: dir, entries: make(map[string]ReviewEntry)}
	_ = s.load()
	return s, nil
}

func (s *ReviewStore) Set(port int, proto string, status ReviewStatus, reviewer, note string) error {
	key := reviewKey(port, proto)
	s.entries[key] = ReviewEntry{
		Port:      port,
		Proto:     proto,
		Status:    status,
		Reviewer:  reviewer,
		Note:      note,
		UpdatedAt: time.Now(),
	}
	return s.save()
}

func (s *ReviewStore) Get(port int, proto string) (ReviewEntry, bool) {
	e, ok := s.entries[reviewKey(port, proto)]
	return e, ok
}

func (s *ReviewStore) Remove(port int, proto string) error {
	key := reviewKey(port, proto)
	if _, ok := s.entries[key]; !ok {
		return fmt.Errorf("review: no entry for %s", key)
	}
	delete(s.entries, key)
	return s.save()
}

func (s *ReviewStore) All() []ReviewEntry {
	out := make([]ReviewEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Port != out[j].Port {
			return out[i].Port < out[j].Port
		}
		return out[i].Proto < out[j].Proto
	})
	return out
}

func (s *ReviewStore) FilterByStatus(status ReviewStatus) []ReviewEntry {
	var out []ReviewEntry
	for _, e := range s.All() {
		if e.Status == status {
			out = append(out, e)
		}
	}
	return out
}

func (s *ReviewStore) save() error {
	data, err := json.MarshalIndent(s.entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, "reviews.json"), data, 0644)
}

func (s *ReviewStore) load() error {
	data, err := os.ReadFile(filepath.Join(s.dir, "reviews.json"))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.entries)
}
