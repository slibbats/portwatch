package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// WatchlistEntry represents a known/trusted listener entry.
type WatchlistEntry struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Process  string `json:"process,omitempty"`
	Note     string `json:"note,omitempty"`
	AddedAt  string `json:"added_at"`
}

// Watchlist holds a collection of trusted listener entries.
type Watchlist struct {
	Entries []WatchlistEntry `json:"entries"`
}

// NewWatchlist returns an empty Watchlist.
func NewWatchlist() *Watchlist {
	return &Watchlist{Entries: []WatchlistEntry{}}
}

// Add inserts a new entry into the watchlist.
func (w *Watchlist) Add(port int, protocol, process, note string) {
	w.Entries = append(w.Entries, WatchlistEntry{
		Port:     port,
		Protocol: protocol,
		Process:  process,
		Note:     note,
		AddedAt:  time.Now().UTC().Format(time.RFC3339),
	})
}

// Contains returns true if the given port/protocol pair is in the watchlist.
func (w *Watchlist) Contains(port int, protocol string) bool {
	for _, e := range w.Entries {
		if e.Port == port && e.Protocol == protocol {
			return true
		}
	}
	return false
}

// SaveWatchlist writes the watchlist to a JSON file.
func SaveWatchlist(path string, w *Watchlist) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("watchlist: create file: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(w)
}

// LoadWatchlist reads a watchlist from a JSON file.
func LoadWatchlist(path string) (*Watchlist, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("watchlist: open file: %w", err)
	}
	defer f.Close()
	var w Watchlist
	if err := json.NewDecoder(f).Decode(&w); err != nil {
		return nil, fmt.Errorf("watchlist: decode: %w", err)
	}
	return &w, nil
}

// FilterUnwatched returns listeners not present in the watchlist.
func (w *Watchlist) FilterUnwatched(listeners []Listener) []Listener {
	var out []Listener
	for _, l := range listeners {
		if !w.Contains(l.Port, l.Protocol) {
			out = append(out, l)
		}
	}
	return out
}
