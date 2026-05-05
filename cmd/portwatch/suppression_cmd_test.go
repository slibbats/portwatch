package main

import (
	"testing"
)

func TestParseSuppressionPortProto_Valid(t *testing.T) {
	port, proto, err := parseSuppressionPortProto("8080/tcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 8080 {
		t.Errorf("port = %d, want 8080", port)
	}
	if proto != "tcp" {
		t.Errorf("proto = %q, want tcp", proto)
	}
}

func TestParseSuppressionPortProto_UDP(t *testing.T) {
	port, proto, err := parseSuppressionPortProto("53/udp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 53 || proto != "udp" {
		t.Errorf("got %d/%s, want 53/udp", port, proto)
	}
}

func TestParseSuppressionPortProto_MissingSlash(t *testing.T) {
	_, _, err := parseSuppressionPortProto("8080")
	if err == nil {
		t.Error("expected error for missing slash")
	}
}

func TestParseSuppressionPortProto_InvalidPort(t *testing.T) {
	_, _, err := parseSuppressionPortProto("abc/tcp")
	if err == nil {
		t.Error("expected error for non-numeric port")
	}
}

func TestParseSuppressionPortProto_EmptyProto(t *testing.T) {
	_, _, err := parseSuppressionPortProto("80/")
	if err == nil {
		t.Error("expected error for empty proto")
	}
}

func TestRunSuppression_UnknownFlag(t *testing.T) {
	err := runSuppression([]string{"--unknown-flag"})
	if err == nil {
		t.Error("expected error for unknown flag")
	}
}

func TestRunSuppression_NoSubcommand(t *testing.T) {
	err := runSuppression([]string{})
	if err == nil {
		t.Error("expected error when no subcommand given")
	}
}

func TestRunSuppression_UnknownSubcommand(t *testing.T) {
	dir := t.TempDir()
	err := runSuppression([]string{"-data-dir", dir, "bogus"})
	if err == nil {
		t.Error("expected error for unknown subcommand")
	}
}

func TestRunSuppression_ListEmpty(t *testing.T) {
	dir := t.TempDir()
	err := runSuppression([]string{"-data-dir", dir, "list"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunSuppression_AddAndList(t *testing.T) {
	dir := t.TempDir()
	if err := runSuppression([]string{"-data-dir", dir, "-reason", "test", "-duration", "1h", "add", "9090/tcp"}); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := runSuppression([]string{"-data-dir", dir, "list"}); err != nil {
		t.Fatalf("list: %v", err)
	}
}

func TestRunSuppression_AddMissingPortArg(t *testing.T) {
	dir := t.TempDir()
	err := runSuppression([]string{"-data-dir", dir, "add"})
	if err == nil {
		t.Error("expected error when port/proto missing")
	}
}
