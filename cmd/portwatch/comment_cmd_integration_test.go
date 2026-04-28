package main

import (
	"os"
	"testing"

	"github.com/user/portwatch/internal/portscanner"
)

func TestCommentStore_SetAndGet_Integration(t *testing.T) {
	dir := t.TempDir()
	store, err := portscanner.NewCommentStore(dir)
	if err != nil {
		t.Fatalf("NewCommentStore: %v", err)
	}
	if err := store.Set(8080, "tcp", "local dev"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	c, ok := store.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected comment")
	}
	if c.Text != "local dev" {
		t.Errorf("expected 'local dev', got %q", c.Text)
	}
}

func TestCommentStore_Remove_Integration(t *testing.T) {
	dir := t.TempDir()
	store, _ := portscanner.NewCommentStore(dir)
	_ = store.Set(22, "tcp", "ssh")
	_ = store.Remove(22, "tcp")
	_, ok := store.Get(22, "tcp")
	if ok {
		t.Fatal("expected comment to be removed")
	}
}

func TestCommentStore_List_Integration(t *testing.T) {
	dir := t.TempDir()
	store, _ := portscanner.NewCommentStore(dir)
	_ = store.Set(80, "tcp", "http")
	_ = store.Set(443, "tcp", "https")
	all := store.All()
	if len(all) != 2 {
		t.Errorf("expected 2 comments, got %d", len(all))
	}
	_ = os.Remove(dir)
}

func TestCommentStore_Persistence_Integration(t *testing.T) {
	dir := t.TempDir()
	s1, _ := portscanner.NewCommentStore(dir)
	_ = s1.Set(3306, "tcp", "mysql")

	s2, err := portscanner.NewCommentStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	c, ok := s2.Get(3306, "tcp")
	if !ok || c.Text != "mysql" {
		t.Error("expected persisted comment after reload")
	}
}
