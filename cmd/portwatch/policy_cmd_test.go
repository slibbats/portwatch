package main

import (
	"testing"
)

func TestParsePolicyPortProto_Valid(t *testing.T) {
	port, proto, err := parsePolicyPortProto("8080/tcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 8080 || proto != "tcp" {
		t.Errorf("got %d/%s", port, proto)
	}
}

func TestParsePolicyPortProto_UDP(t *testing.T) {
	port, proto, err := parsePolicyPortProto("53/udp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 53 || proto != "udp" {
		t.Errorf("got %d/%s", port, proto)
	}
}

func TestParsePolicyPortProto_MissingSlash(t *testing.T) {
	_, _, err := parsePolicyPortProto("8080")
	if err == nil {
		t.Fatal("expected error for missing slash")
	}
}

func TestParsePolicyPortProto_InvalidPort(t *testing.T) {
	_, _, err := parsePolicyPortProto("abc/tcp")
	if err == nil {
		t.Fatal("expected error for non-numeric port")
	}
}

func TestParsePolicyPortProto_EmptyProto(t *testing.T) {
	_, _, err := parsePolicyPortProto("80/")
	if err == nil {
		t.Fatal("expected error for empty proto")
	}
}
