package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type EscalationPolicy struct {
	Port     int    `json:"port"`
	Proto    string `json:"proto"`
	Contact  string `json:"contact"`
	Channel  string `json:"channel"`
	MinLevel string `json:"min_level"`
}

func escalationKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, strings.ToLower(proto))
}

type EscalationStore struct {
	dir      string
	policies map[string]EscalationPolicy
}

func NewEscalationStore(dir string) (*EscalationStore, error) {
	s := &EscalationStore{dir: dir, policies: make(map[string]EscalationPolicy)}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	_ = s.load()
	return s, nil
}

func (s *EscalationStore) Set(port int, proto, contact, channel, minLevel string) error {
	s.policies[escalationKey(port, proto)] = EscalationPolicy{
		Port: port, Proto: strings.ToLower(proto),
		Contact: contact, Channel: channel, MinLevel: minLevel,
	}
	return s.save()
}

func (s *EscalationStore) Get(port int, proto string) (EscalationPolicy, bool) {
	p, ok := s.policies[escalationKey(port, proto)]
	return p, ok
}

func (s *EscalationStore) Remove(port int, proto string) error {
	key := escalationKey(port, proto)
	if _, ok := s.policies[key]; !ok {
		return fmt.Errorf("no escalation policy for %s", key)
	}
	delete(s.policies, key)
	return s.save()
}

func (s *EscalationStore) All() []EscalationPolicy {
	out := make([]EscalationPolicy, 0, len(s.policies))
	for _, p := range s.policies {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Port != out[j].Port {
			return out[i].Port < out[j].Port
		}
		return out[i].Proto < out[j].Proto
	})
	return out
}

func (s *EscalationStore) save() error {
	path := filepath.Join(s.dir, "escalation.json")
	data, err := json.MarshalIndent(s.policies, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (s *EscalationStore) load() error {
	path := filepath.Join(s.dir, "escalation.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.policies)
}

func parseEscalationPort(raw string) (int, string, error) {
	parts := strings.SplitN(raw, "/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return 0, "", fmt.Errorf("expected port/proto, got %q", raw)
	}
	port, err := strconv.Atoi(parts[0])
	if err != nil || port < 1 || port > 65535 {
		return 0, "", fmt.Errorf("invalid port %q", parts[0])
	}
	return port, parts[1], nil
}
