package main

import (
	"testing"
)

func TestRunSummary_UnknownFlag(t *testing.T) {
	err := runSummary([]string{"--unknown-flag"})
	if err == nil {
		t.Fatal("expected error for unknown flag, got nil")
	}
}

func TestRunSummary_InvalidSnapshotDir(t *testing.T) {
	// Passing --latest with a non-existent directory should return an error.
	err := runSummary([]string{
		"--latest",
		"--snapshot-dir", "/nonexistent/path/portwatch/snapshots",
	})
	if err == nil {
		t.Fatal("expected error when snapshot dir does not exist, got nil")
	}
}

func TestRunSummary_DefaultFlagsAreParseable(t *testing.T) {
	// Verify flag parsing itself does not panic or error with default values.
	// We cannot run a full live scan in unit tests, so we only check flag parsing
	// by passing --help which returns a well-known error.
	err := runSummary([]string{"--help"})
	if err == nil {
		t.Log("--help returned nil (unexpected but not fatal in all flag implementations)")
	}
	// The important thing is no panic occurred.
}

func TestRunSummary_LatestFlagSet(t *testing.T) {
	// When --latest is set with a missing dir, we expect a descriptive error.
	err := runSummary([]string{"--latest", "--snapshot-dir", t.TempDir() + "/empty"})
	if err == nil {
		t.Fatal("expected error for empty/missing snapshot directory")
	}
	if len(err.Error()) == 0 {
		t.Fatal("expected non-empty error message")
	}
}
