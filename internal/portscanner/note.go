package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Note struct {
	Port      int       `json:"port"`
	Protocol  string    `json:"protocol"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type noteKey struct {
	port     int
	protocol string
}

type NoteStore struct {
	mu    sync.RWMutex
	notes map[noteKey]*Note
	path  string
}

func NewNoteStore(path string) (*NoteStore, error) {
	ns := &NoteStore{
		notes: make(map[noteKey]*Note),
		path:  path,
	}
	if err := ns.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load notes: %w", err)
	}
	return ns, nil
}

func (ns *NoteStore) Set(port int, protocol, text string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	k := noteKey{port: port, protocol: protocol}
	now := time.Now().UTC()
	if existing, ok := ns.notes[k]; ok {
		existing.Text = text
		existing.UpdatedAt = now
		return
	}
	ns.notes[k] = &Note{Port: port, Protocol: protocol, Text: text, CreatedAt: now, UpdatedAt: now}
}

func (ns *NoteStore) Get(port int, protocol string) (*Note, bool) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	n, ok := ns.notes[noteKey{port: port, protocol: protocol}]
	if !ok {
		return nil, false
	}
	copy := *n
	return &copy, true
}

func (ns *NoteStore) Remove(port int, protocol string) bool {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	k := noteKey{port: port, protocol: protocol}
	if _, ok := ns.notes[k]; !ok {
		return false
	}
	delete(ns.notes, k)
	return true
}

func (ns *NoteStore) All() []*Note {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	out := make([]*Note, 0, len(ns.notes))
	for _, n := range ns.notes {
		copy := *n
		out = append(out, &copy)
	}
	return out
}

func (ns *NoteStore) Save() error {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	if err := os.MkdirAll(filepath.Dir(ns.path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(ns.path)
	if err != nil {
		return err
	}
	defer f.Close()
	notes := make([]*Note, 0, len(ns.notes))
	for _, n := range ns.notes {
		notes = append(notes, n)
	}
	return json.NewEncoder(f).Encode(notes)
}

func (ns *NoteStore) load() error {
	f, err := os.Open(ns.path)
	if err != nil {
		return err
	}
	defer f.Close()
	var notes []*Note
	if err := json.NewDecoder(f).Decode(&notes); err != nil {
		return err
	}
	for _, n := range notes {
		ns.notes[noteKey{port: n.Port, protocol: n.Protocol}] = n
	}
	return nil
}
