package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Policy struct {
	Port     int    `json:"port"`
	Proto    string `json:"proto"`
	Action   string `json:"action"` // "allow" or "deny"
	Reason   string `json:"reason"`
}

func policyKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

type PolicyStore struct {
	dir      string
	policies map[string]Policy
}

func NewPolicyStore(dir string) (*PolicyStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("policy store: mkdir: %w", err)
	}
	ps := &PolicyStore{dir: dir, policies: make(map[string]Policy)}
	if err := ps.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return ps, nil
}

func (ps *PolicyStore) Set(port int, proto, action, reason string) error {
	ps.policies[policyKey(port, proto)] = Policy{Port: port, Proto: proto, Action: action, Reason: reason}
	return ps.save()
}

func (ps *PolicyStore) Get(port int, proto string) (Policy, bool) {
	p, ok := ps.policies[policyKey(port, proto)]
	return p, ok
}

func (ps *PolicyStore) Remove(port int, proto string) error {
	key := policyKey(port, proto)
	if _, ok := ps.policies[key]; !ok {
		return fmt.Errorf("policy not found: %s", key)
	}
	delete(ps.policies, key)
	return ps.save()
}

func (ps *PolicyStore) All() []Policy {
	out := make([]Policy, 0, len(ps.policies))
	for _, p := range ps.policies {
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

func (ps *PolicyStore) path() string {
	return filepath.Join(ps.dir, "policies.json")
}

func (ps *PolicyStore) save() error {
	data, err := json.MarshalIndent(ps.policies, "", "  ")
	if err != nil {
		return fmt.Errorf("policy store: marshal: %w", err)
	}
	return os.WriteFile(ps.path(), data, 0644)
}

func (ps *PolicyStore) load() error {
	data, err := os.ReadFile(ps.path())
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &ps.policies)
}
