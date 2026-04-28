package main

import (
	"testing"
)

func TestParsePortProto_Valid(t *testing.T) {
	port, proto, err := parsePortProto("8080/tcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 8080 {
		t.Errorf("expected port 8080, got %d", port)
	}
	if proto != "tcp" {
		t.Errorf("expected proto 'tcp', got %q", proto)
	}
}

func TestParsePortProto_UDP(t *testing.T) {
	port, proto, err := parsePortProto("53/udp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 53 || proto != "udp" {
		t.Errorf("unexpected result: %d/%s", port, proto)
	}
}

func TestParsePortProto_MissingSlash(t *testing.T) {
	_, _, err := parsePortProto("8080")
	if err == nil {
		t.Fatal("expected error for missing slash")
	}
}

func TestParsePortProto_InvalidPort(t *testing.T) {
	_, _, err := parsePortProto("abc/tcp")
	if err == nil {
		t.Fatal("expected error for non-numeric port")
	}
}

func TestParsePortProto_EmptyProto(t *testing.T) {
	port, proto, err := parsePortProto("443/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 443 {
		t.Errorf("expected 443, got %d", port)
	}
	if proto != "" {
		t.Errorf("expected empty proto, got %q", proto)
	}
}
