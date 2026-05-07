package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type RetentionPolicy struct {
	MaxAgeDays int `json:"max_age_days"`
	MaxCount   int `json:"max_count"`
}

type RetentionResult struct {
	Pruned  int
	Remaining int
	Errors  []error
}

func retentionKey(dir string) string {
	return filepath.Join(dir, "retention_policy.json")
}

func DefaultRetentionPolicy() RetentionPolicy {
	return RetentionPolicy{
		MaxAgeDays: 30,
		MaxCount:   100,
	}
}

func SaveRetentionPolicy(dir string, p RetentionPolicy) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create retention dir: %w", err)
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal retention policy: %w", err)
	}
	return os.WriteFile(retentionKey(dir), data, 0644)
}

func LoadRetentionPolicy(dir string) (RetentionPolicy, error) {
	data, err := os.ReadFile(retentionKey(dir))
	if os.IsNotExist(err) {
		return DefaultRetentionPolicy(), nil
	}
	if err != nil {
		return RetentionPolicy{}, fmt.Errorf("read retention policy: %w", err)
	}
	var p RetentionPolicy
	if err := json.Unmarshal(data, &p); err != nil {
		return RetentionPolicy{}, fmt.Errorf("unmarshal retention policy: %w", err)
	}
	return p, nil
}

func ApplyRetention(snapshotDir string, policy RetentionPolicy) (RetentionResult, error) {
	entries, err := os.ReadDir(snapshotDir)
	if os.IsNotExist(err) {
		return RetentionResult{}, nil
	}
	if err != nil {
		return RetentionResult{}, fmt.Errorf("read snapshot dir: %w", err)
	}

	var files []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" && e.Name() != "retention_policy.json" {
			files = append(files, e)
		}
	}

	sort.Slice(files, func(i, j int) bool {
		ii, _ := files[i].Info()
		jj, _ := files[j].Info()
		if ii == nil || jj == nil {
			return false
		}
		return ii.ModTime().Before(jj.ModTime())
	})

	result := RetentionResult{Remaining: len(files)}
	cutoff := time.Now().AddDate(0, 0, -policy.MaxAgeDays)

	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			result.Errors = append(result.Errors, err)
			continue
		}
		shouldPrune := info.ModTime().Before(cutoff) || (policy.MaxCount > 0 && result.Remaining > policy.MaxCount)
		if shouldPrune {
			path := filepath.Join(snapshotDir, f.Name())
			if err := os.Remove(path); err != nil {
				result.Errors = append(result.Errors, err)
			} else {
				result.Pruned++
				result.Remaining--
			}
		}
	}
	return result, nil
}
