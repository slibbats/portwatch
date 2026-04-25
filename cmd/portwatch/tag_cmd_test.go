package main

import (
	"path/filepath"
	"testing"
)

func TestRunTag_UnknownFlag(t *testing.T) {
	err := runTag([]string{"--notaflag"})
	if err == nil {
		t.Error("expected error for unknown flag")
	}
}

func TestRunTag_MissingArgs(t *testing.T) {
	dir := t.TempDir()
	tagFile := filepath.Join(dir, "tags.json")
	err := runTag([]string{"-file", tagFile})
	if err == nil {
		t.Error("expected error when no port/label provided")
	}
}

func TestRunTag_InvalidPortFormat(t *testing.T) {
	dir := t.TempDir()
	tagFile := filepath.Join(dir, "tags.json")
	err := runTag([]string{"-file", tagFile, "badformat", "label"})
	if err == nil {
		t.Error("expected error for bad port format")
	}
}

func TestRunTag_AddAndList(t *testing.T) {
	dir := t.TempDir()
	tagFile := filepath.Join(dir, "tags.json")

	if err := runTag([]string{"-file", tagFile, "8080/tcp", "dev-server"}); err != nil {
		t.Fatalf("add tag: %v", err)
	}

	if err := runTag([]string{"-file", tagFile, "-list"}); err != nil {
		t.Fatalf("list tags: %v", err)
	}
}

func TestRunTag_Remove(t *testing.T) {
	dir := t.TempDir()
	tagFile := filepath.Join(dir, "tags.json")

	if err := runTag([]string{"-file", tagFile, "9090/tcp", "test"}); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := runTag([]string{"-file", tagFile, "-remove", "9090/tcp", "test"}); err != nil {
		t.Fatalf("remove: %v", err)
	}
}

func TestRunTag_ListEmpty(t *testing.T) {
	dir := t.TempDir()
	tagFile := filepath.Join(dir, "tags.json")

	if err := runTag([]string{"-file", tagFile, "-list"}); err != nil {
		t.Fatalf("list empty: %v", err)
	}
}

func TestRunTag_WithNote(t *testing.T) {
	dir := t.TempDir()
	tagFile := filepath.Join(dir, "tags.json")

	err := runTag([]string{"-file", tagFile, "-note", "internal only", "5432/tcp", "postgres"})
	if err != nil {
		t.Fatalf("add with note: %v", err)
	}
}
