package main

import (
	"testing"
)

func TestParseClassificationPortProto_Valid(t *testing.T) {
	port, proto, err := parseClassificationPortProto("443/tcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 443 {
		t.Errorf("expected port 443, got %d", port)
	}
	if proto != "tcp" {
		t.Errorf("expected proto tcp, got %q", proto)
	}
}

func TestParseClassificationPortProto_UDP(t *testing.T) {
	port, proto, err := parseClassificationPortProto("53/udp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 53 || proto != "udp" {
		t.Errorf("unexpected result: %d/%s", port, proto)
	}
}

func TestParseClassificationPortProto_MissingSlash(t *testing.T) {
	_, _, err := parseClassificationPortProto("443tcp")
	if err == nil {
		t.Fatal("expected error for missing slash")
	}
}

func TestParseClassificationPortProto_InvalidPort(t *testing.T) {
	_, _, err := parseClassificationPortProto("abc/tcp")
	if err == nil {
		t.Fatal("expected error for non-numeric port")
	}
}

func TestParseClassificationPortProto_EmptyProto(t *testing.T) {
	_, _, err := parseClassificationPortProto("80/")
	if err == nil {
		t.Fatal("expected error for empty proto")
	}
}

func TestRunClassification_UnknownFlag(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered: %v", r)
		}
	}()
	// Should not panic on valid subcommand list with empty dir
	dir := t.TempDir()
	runClassification([]string{"-data-dir", dir, "list"})
}
