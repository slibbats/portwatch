package main

import (
	"testing"
)

func TestParseAlertPortProto_Valid(t *testing.T) {
	port, proto, err := parseAlertPortProto("8080/tcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 8080 || proto != "tcp" {
		t.Errorf("got port=%d proto=%s", port, proto)
	}
}

func TestParseAlertPortProto_UDP(t *testing.T) {
	port, proto, err := parseAlertPortProto("53/udp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 53 || proto != "udp" {
		t.Errorf("got port=%d proto=%s", port, proto)
	}
}

func TestParseAlertPortProto_MissingSlash(t *testing.T) {
	_, _, err := parseAlertPortProto("8080")
	if err == nil {
		t.Fatal("expected error for missing slash")
	}
}

func TestParseAlertPortProto_InvalidPort(t *testing.T) {
	_, _, err := parseAlertPortProto("notaport/tcp")
	if err == nil {
		t.Fatal("expected error for non-numeric port")
	}
}

func TestParseAlertPortProto_EmptyProto(t *testing.T) {
	_, _, err := parseAlertPortProto("80/")
	if err == nil {
		t.Fatal("expected error for empty proto")
	}
}

func TestRunAlertRule_UnknownFlag(t *testing.T) {
	err := runAlertRule([]string{"--no-such-flag"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

func TestRunAlertRule_NoSubcommand(t *testing.T) {
	err := runAlertRule([]string{})
	if err == nil {
		t.Fatal("expected error when no subcommand given")
	}
}

func TestRunAlertRule_UnknownSubcommand(t *testing.T) {
	dir := t.TempDir()
	err := runAlertRule([]string{"-data-dir", dir, "bogus"})
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
}

func TestRunAlertRule_AddAndList(t *testing.T) {
	dir := t.TempDir()
	if err := runAlertRule([]string{"-data-dir", dir, "add", "3306/tcp", "high", "mysql exposed"}); err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if err := runAlertRule([]string{"-data-dir", dir, "list"}); err != nil {
		t.Fatalf("list failed: %v", err)
	}
}

func TestRunAlertRule_Remove(t *testing.T) {
	dir := t.TempDir()
	_ = runAlertRule([]string{"-data-dir", dir, "add", "22/tcp", "critical"})
	if err := runAlertRule([]string{"-data-dir", dir, "remove", "22/tcp"}); err != nil {
		t.Fatalf("remove failed: %v", err)
	}
}

func TestRunAlertRule_Remove_NotFound(t *testing.T) {
	dir := t.TempDir()
	err := runAlertRule([]string{"-data-dir", dir, "remove", "9999/tcp"})
	if err == nil {
		t.Fatal("expected error when removing non-existent rule")
	}
}
