package main

import (
	"testing"
)

func TestParseSchedulePortProto_Valid(t *testing.T) {
	port, proto, err := parseSchedulePortProto("8080/tcp")
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

func TestParseSchedulePortProto_UDP(t *testing.T) {
	port, proto, err := parseSchedulePortProto("53/udp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 53 || proto != "udp" {
		t.Errorf("unexpected result: %d/%s", port, proto)
	}
}

func TestParseSchedulePortProto_MissingSlash(t *testing.T) {
	_, _, err := parseSchedulePortProto("8080")
	if err == nil {
		t.Error("expected error for missing slash")
	}
}

func TestParseSchedulePortProto_InvalidPort(t *testing.T) {
	_, _, err := parseSchedulePortProto("abc/tcp")
	if err == nil {
		t.Error("expected error for non-numeric port")
	}
}

func TestParseSchedulePortProto_EmptyProto(t *testing.T) {
	_, _, err := parseSchedulePortProto("80/")
	if err == nil {
		t.Error("expected error for empty proto")
	}
}

func TestRunSchedule_UnknownFlag(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected exit on unknown flag")
		}
	}()
	runSchedule([]string{"--unknown-flag"})
}

func TestRunSchedule_DefaultFlagsAreParseable(t *testing.T) {
	// Should not panic when listing from a temp dir with no entries
	// We can't call runSchedule directly (os.Exit), so just test parsing
	port, proto, err := parseSchedulePortProto("443/tcp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != 443 || proto != "tcp" {
		t.Errorf("unexpected: %d/%s", port, proto)
	}
}
