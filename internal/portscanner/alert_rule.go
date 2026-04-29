package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type AlertRule struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type AlertRuleStore struct {
	rules map[string]AlertRule
	path  string
}

func alertRuleKey(port int, proto string) string {
	return fmt.Sprintf("%d/%s", port, proto)
}

func NewAlertRuleStore(dir string) (*AlertRuleStore, error) {
	s := &AlertRuleStore{
		rules: make(map[string]AlertRule),
		path:  filepath.Join(dir, "alert_rules.json"),
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *AlertRuleStore) Set(rule AlertRule) {
	s.rules[alertRuleKey(rule.Port, rule.Protocol)] = rule
}

func (s *AlertRuleStore) Get(port int, proto string) (AlertRule, bool) {
	r, ok := s.rules[alertRuleKey(port, proto)]
	return r, ok
}

func (s *AlertRuleStore) Remove(port int, proto string) bool {
	key := alertRuleKey(port, proto)
	if _, ok := s.rules[key]; !ok {
		return false
	}
	delete(s.rules, key)
	return true
}

func (s *AlertRuleStore) All() []AlertRule {
	out := make([]AlertRule, 0, len(s.rules))
	for _, r := range s.rules {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Port != out[j].Port {
			return out[i].Port < out[j].Port
		}
		return out[i].Protocol < out[j].Protocol
	})
	return out
}

func (s *AlertRuleStore) Save() error {
	data, err := json.MarshalIndent(s.rules, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *AlertRuleStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.rules)
}
