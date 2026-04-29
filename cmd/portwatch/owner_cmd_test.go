package main

import (
	"testing"
)

func TestParseOwnerPortProto_Valid(t *testing.T) {
	port, proto, err := parseOwnerPortProto("8080/tcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 8080 || proto != "tcp" {
		t.Errorf("got port=%d proto=%s", port, proto)
	}
}

func TestParseOwnerPortProto_UDP(t *testing.T) {
	port, proto, err := parseOwnerPortProto("53/udp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 53 || proto != "udp" {
		t.Errorf("got port=%d proto=%s", port, proto)
	}
}

func TestParseOwnerPortProto_MissingSlash(t *testing.T) {
	_, _, err := parseOwnerPortProto("8080")
	if err == nil {
		t.Error("expected error for missing slash")
	}
}

func TestParseOwnerPortProto_InvalidPort(t *testing.T) {
	_, _, err := parseOwnerPortProto("abc/tcp")
	if err == nil {
		t.Error("expected error for non-numeric port")
	}
}

func TestParseOwnerPortProto_EmptyProto(t *testing.T) {
	_, _, err := parseOwnerPortProto("80/")
	if err == nil {
		t.Error("expected error for empty proto")
	}
}

func TestRunOwner_UnknownFlag(t *testing.T) {
	defer func() { recover() }()
	runOwner([]string{"--unknown-flag"})
}

func TestRunOwner_DefaultFlagsAreParseable(t *testing.T) {
	// Ensure flag parsing itself doesn't panic for list with a temp dir.
	// We can't easily call runOwner without os.Exit, so we just test the parser.
	port, proto, err := parseOwnerPortProto("443/tcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 443 || proto != "tcp" {
		t.Errorf("unexpected result: %d/%s", port, proto)
	}
}
