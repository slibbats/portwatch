package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type IncidentSeverity string

const (
	IncidentLow      IncidentSeverity = "low"
	IncidentMedium   IncidentSeverity = "medium"
	IncidentHigh     IncidentSeverity = "high"
	IncidentCritical IncidentSeverity = "critical"
)

type Incident struct {
	ID        string           `json:"id"`
	Port      int              `json:"port"`
	Proto     string           `json:"proto"`
	Severity  IncidentSeverity `json:"severity"`
	Message   string           `json:"message"`
	CreatedAt time.Time        `json:"created_at"`
	ResolvedAt *time.Time      `json:"resolved_at,omitempty"`
}

type IncidentStore struct {
	dir string
}

func incidentKey(port int, proto string) string {
	return fmt.Sprintf("%d_%s", port, proto)
}

func NewIncidentStore(dir string) *IncidentStore {
	return &IncidentStore{dir: dir}
}

func (s *IncidentStore) Open(port int, proto string, severity IncidentSeverity, message string) (*Incident, error) {
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return nil, fmt.Errorf("incident: mkdir: %w", err)
	}
	inc := &Incident{
		ID:        fmt.Sprintf("%s_%d", incidentKey(port, proto), time.Now().UnixNano()),
		Port:      port,
		Proto:     proto,
		Severity:  severity,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	}
	path := filepath.Join(s.dir, inc.ID+".json")
	data, err := json.MarshalIndent(inc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("incident: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return nil, fmt.Errorf("incident: write: %w", err)
	}
	return inc, nil
}

func (s *IncidentStore) Resolve(id string) error {
	path := filepath.Join(s.dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("incident: read: %w", err)
	}
	var inc Incident
	if err := json.Unmarshal(data, &inc); err != nil {
		return fmt.Errorf("incident: unmarshal: %w", err)
	}
	now := time.Now().UTC()
	inc.ResolvedAt = &now
	updated, err := json.MarshalIndent(inc, "", "  ")
	if err != nil {
		return fmt.Errorf("incident: marshal: %w", err)
	}
	return os.WriteFile(path, updated, 0644)
}

func (s *IncidentStore) All() ([]Incident, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("incident: readdir: %w", err)
	}
	var incidents []Incident
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			continue
		}
		var inc Incident
		if err := json.Unmarshal(data, &inc); err != nil {
			continue
		}
		incidents = append(incidents, inc)
	}
	sort.Slice(incidents, func(i, j int) bool {
		return incidents[i].CreatedAt.Before(incidents[j].CreatedAt)
	})
	return incidents, nil
}

func (s *IncidentStore) OpenOnly() ([]Incident, error) {
	all, err := s.All()
	if err != nil {
		return nil, err
	}
	var open []Incident
	for _, inc := range all {
		if inc.ResolvedAt == nil {
			open = append(open, inc)
		}
	}
	return open, nil
}
