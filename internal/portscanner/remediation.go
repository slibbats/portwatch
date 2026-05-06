package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type RemediationAction struct {
	Port     int    `json:"port"`
	Proto    string `json:"proto"`
	Action   string `json:"action"`
	Script   string `json:"script,omitempty"`
	RunAs    string `json:"run_as,omitempty"`
}

func remediationKey(port int, proto string) string {
	return fmt.Sprintf("%d_%s", port, proto)
}

type RemediationStore struct {
	dir  string
	items map[string]RemediationAction
}

func NewRemediationStore(dir string) (*RemediationStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("remediation: mkdir %s: %w", dir, err)
	}
	s := &RemediationStore{dir: dir, items: make(map[string]RemediationAction)}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *RemediationStore) Set(r RemediationAction) error {
	s.items[remediationKey(r.Port, r.Proto)] = r
	return s.save()
}

func (s *RemediationStore) Get(port int, proto string) (RemediationAction, bool) {
	r, ok := s.items[remediationKey(port, proto)]
	return r, ok
}

func (s *RemediationStore) Remove(port int, proto string) error {
	key := remediationKey(port, proto)
	if _, ok := s.items[key]; !ok {
		return fmt.Errorf("remediation: no entry for %d/%s", port, proto)
	}
	delete(s.items, key)
	return s.save()
}

func (s *RemediationStore) All() []RemediationAction {
	out := make([]RemediationAction, 0, len(s.items))
	for _, v := range s.items {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Port != out[j].Port {
			return out[i].Port < out[j].Port
		}
		return out[i].Proto < out[j].Proto
	})
	return out
}

func (s *RemediationStore) save() error {
	path := filepath.Join(s.dir, "remediation.json")
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *RemediationStore) load() error {
	path := filepath.Join(s.dir, "remediation.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.items)
}
