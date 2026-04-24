package alerting

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds the alerting configuration loaded from a JSON file.
type Config struct {
	// AllowedPorts is the list of ports considered safe/expected.
	AllowedPorts []uint16 `json:"allowed_ports"`
	// PollIntervalSeconds is how often (in seconds) to scan for listeners.
	PollIntervalSeconds int `json:"poll_interval_seconds"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		AllowedPorts:        []uint16{22, 80, 443},
		PollIntervalSeconds: 30,
	}
}

// LoadConfig reads and parses a JSON config file from the given path.
// If the path is empty, DefaultConfig is returned.
func LoadConfig(path string) (Config, error) {
	if path == "" {
		return DefaultConfig(), nil
	}

	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("alerting: open config %q: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("alerting: decode config %q: %w", path, err)
	}

	if cfg.PollIntervalSeconds <= 0 {
		cfg.PollIntervalSeconds = DefaultConfig().PollIntervalSeconds
	}

	return cfg, nil
}
