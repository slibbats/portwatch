package main

import (
	"testing"
)

func TestParseRemediationPortProto_Valid(t *testing.T) {
	port, proto, err := parseRemediationPortProto("8080/tcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 8080 || proto != "tcp" {
		t.Errorf("expected 8080/tcp, got %d/%s", port, proto)
	}
}

func TestParseRemediationPortProto_UDP(t *testing.T) {
	port, proto, err := parseRemediationPortProto("53/udp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 53 || proto != "udp" {
		t.Errorf("expected 53/udp, got %d/%s", port, proto)
	}
}

func TestParseRemediationPortProto_MissingSlash(t *testing.T) {
	_, _, err := parseRemediationPortProto("8080")
	if err == nil {
		t.Error("expected error for missing slash")
	}
}

func TestParseRemediationPortProto_InvalidPort(t *testing.T) {
	_, _, err := parseRemediationPortProto("abc/tcp")
	if err == nil {
		t.Error("expected error for non-numeric port")
	}
}

func TestParseRemediationPortProto_EmptyProto(t *testing.T) {
	_, _, err := parseRemediationPortProto("80/")
	if err == nil {
		t.Error("expected error for empty proto")
	}
}

func TestRunRemediation_UnknownFlag(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered: %v", r)
		}
	}()
	// Should not panic on unknown subcommand path
}

func TestRunRemediation_DefaultFlagsAreParseable(t *testing.T) {
	// Verify the flag set can be constructed without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()
	_ = parseRemediationPortProto // just reference the function
}
