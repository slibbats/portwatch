package main

import (
	"testing"
)

func TestParseDependencyPortProto_Valid(t *testing.T) {
	port, proto, err := parseDependencyPortProto("8080/tcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 8080 || proto != "tcp" {
		t.Errorf("got %d/%s", port, proto)
	}
}

func TestParseDependencyPortProto_UDP(t *testing.T) {
	port, proto, err := parseDependencyPortProto("53/udp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 53 || proto != "udp" {
		t.Errorf("got %d/%s", port, proto)
	}
}

func TestParseDependencyPortProto_MissingSlash(t *testing.T) {
	_, _, err := parseDependencyPortProto("8080")
	if err == nil {
		t.Error("expected error for missing slash")
	}
}

func TestParseDependencyPortProto_InvalidPort(t *testing.T) {
	_, _, err := parseDependencyPortProto("abc/tcp")
	if err == nil {
		t.Error("expected error for non-numeric port")
	}
}

func TestParseDependencyPortProto_EmptyProto(t *testing.T) {
	_, _, err := parseDependencyPortProto("8080/")
	if err == nil {
		t.Error("expected error for empty proto")
	}
}

func TestRunDependency_UnknownFlag(t *testing.T) {
	err := runDependency([]string{"--unknown-flag"})
	if err == nil {
		t.Error("expected error for unknown flag")
	}
}

func TestRunDependency_MissingSubcommand(t *testing.T) {
	err := runDependency([]string{})
	if err == nil {
		t.Error("expected error when no subcommand given")
	}
}

func TestRunDependency_UnknownSubcommand(t *testing.T) {
	dir := t.TempDir()
	err := runDependency([]string{"--data-dir", dir, "bogus"})
	if err == nil {
		t.Error("expected error for unknown subcommand")
	}
}
