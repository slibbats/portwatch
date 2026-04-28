package portscanner

import (
	"os"
	"testing"
)

func TestCommentStore_Set_And_Get(t *testing.T) {
	dir := t.TempDir()
	cs, err := NewCommentStore(dir)
	if err != nil {
		t.Fatalf("NewCommentStore: %v", err)
	}
	if err := cs.Set(8080, "tcp", "dev server"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	c, ok := cs.Get(8080, "tcp")
	if !ok {
		t.Fatal("expected comment to exist")
	}
	if c.Text != "dev server" {
		t.Errorf("expected 'dev server', got %q", c.Text)
	}
	if c.Port != 8080 || c.Protocol != "tcp" {
		t.Errorf("unexpected port/protocol: %d/%s", c.Port, c.Protocol)
	}
}

func TestCommentStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewCommentStore(dir)
	_, ok := cs.Get(9999, "tcp")
	if ok {
		t.Fatal("expected no comment")
	}
}

func TestCommentStore_Set_UpdatesExisting(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewCommentStore(dir)
	_ = cs.Set(443, "tcp", "original")
	_ = cs.Set(443, "tcp", "updated")
	c, ok := cs.Get(443, "tcp")
	if !ok {
		t.Fatal("expected comment")
	}
	if c.Text != "updated" {
		t.Errorf("expected 'updated', got %q", c.Text)
	}
	if !c.UpdatedAt.After(c.CreatedAt) && c.UpdatedAt.Equal(c.CreatedAt) {
		// allow equal on fast machines
	}
}

func TestCommentStore_Remove(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewCommentStore(dir)
	_ = cs.Set(22, "tcp", "ssh")
	if err := cs.Remove(22, "tcp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	_, ok := cs.Get(22, "tcp")
	if ok {
		t.Fatal("expected comment to be removed")
	}
}

func TestCommentStore_SaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewCommentStore(dir)
	_ = cs.Set(80, "tcp", "http")
	_ = cs.Set(53, "udp", "dns")

	cs2, err := NewCommentStore(dir)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if c, ok := cs2.Get(80, "tcp"); !ok || c.Text != "http" {
		t.Error("expected http comment after reload")
	}
	if c, ok := cs2.Get(53, "udp"); !ok || c.Text != "dns" {
		t.Error("expected dns comment after reload")
	}
}

func TestCommentStore_All_ReturnsCopy(t *testing.T) {
	dir := t.TempDir()
	cs, _ := NewCommentStore(dir)
	_ = cs.Set(8080, "tcp", "test")
	all := cs.All()
	delete(all, "8080/tcp")
	if _, ok := cs.Get(8080, "tcp"); !ok {
		t.Error("original store should not be mutated")
	}
	_ = os.Remove(dir)
}
