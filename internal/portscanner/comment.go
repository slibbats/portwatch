package portscanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Comment struct {
	Port      int       `json:"port"`
	Protocol  string    `json:"protocol"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CommentStore struct {
	mu       sync.RWMutex
	comments map[string]*Comment
	path     string
}

func commentKey(port int, protocol string) string {
	return fmt.Sprintf("%d/%s", port, protocol)
}

func NewCommentStore(dir string) (*CommentStore, error) {
	cs := &CommentStore{
		comments: make(map[string]*Comment),
		path:     filepath.Join(dir, "comments.json"),
	}
	if err := cs.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return cs, nil
}

func (cs *CommentStore) Set(port int, protocol, text string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	key := commentKey(port, protocol)
	now := time.Now()
	if existing, ok := cs.comments[key]; ok {
		existing.Text = text
		existing.UpdatedAt = now
	} else {
		cs.comments[key] = &Comment{
			Port:      port,
			Protocol:  protocol,
			Text:      text,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	return cs.save()
}

func (cs *CommentStore) Get(port int, protocol string) (*Comment, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	c, ok := cs.comments[commentKey(port, protocol)]
	return c, ok
}

func (cs *CommentStore) Remove(port int, protocol string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.comments, commentKey(port, protocol))
	return cs.save()
}

func (cs *CommentStore) All() map[string]*Comment {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	copy := make(map[string]*Comment, len(cs.comments))
	for k, v := range cs.comments {
		copy[k] = v
	}
	return copy
}

func (cs *CommentStore) save() error {
	data, err := json.MarshalIndent(cs.comments, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cs.path, data, 0644)
}

func (cs *CommentStore) load() error {
	data, err := os.ReadFile(cs.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &cs.comments)
}
