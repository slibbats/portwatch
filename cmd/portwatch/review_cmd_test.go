package main

import (
	"testing"
)

func TestParseReviewPortProto_Valid(t *testing.T) {
	port, proto, err := parseReviewPortProto("8080/tcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 8080 || proto != "tcp" {
		t.Errorf("expected 8080/tcp, got %d/%s", port, proto)
	}
}

func TestParseReviewPortProto_UDP(t *testing.T) {
	port, proto, err := parseReviewPortProto("53/udp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 53 || proto != "udp" {
		t.Errorf("expected 53/udp, got %d/%s", port, proto)
	}
}

func TestParseReviewPortProto_MissingSlash(t *testing.T) {
	_, _, err := parseReviewPortProto("8080")
	if err == nil {
		t.Error("expected error for missing slash")
	}
}

func TestParseReviewPortProto_InvalidPort(t *testing.T) {
	_, _, err := parseReviewPortProto("abc/tcp")
	if err == nil {
		t.Error("expected error for non-numeric port")
	}
}

func TestParseReviewPortProto_EmptyProto(t *testing.T) {
	_, _, err := parseReviewPortProto("80/")
	if err == nil {
		t.Error("expected error for empty proto")
	}
}

func TestRunReview_UnknownFlag(t *testing.T) {
	defer func() { recover() }()
	runReview([]string{"--unknown-flag"})
}

func TestRunReview_DefaultFlagsAreParseable(t *testing.T) {
	// just ensure flag parsing does not panic with valid list subcommand
	dir := t.TempDir()
	runReview([]string{"-data-dir", dir, "list"})
}
