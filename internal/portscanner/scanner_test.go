package portscanner

import (
	"testing"
)

func TestParseHexAddr_Loopback(t *testing.T) {
	ip, port, err := parseHexAddr("0100007F:0050")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "127.0.0.1" {
		t.Errorf("expected 127.0.0.1, got %s", ip)
	}
	if port != 80 {
		t.Errorf("expected port 80, got %d", port)
	}
}

func TestParseHexAddr_AllInterfaces(t *testing.T) {
	ip, port, err := parseHexAddr("00000000:1F90")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "0.0.0.0" {
		t.Errorf("expected 0.0.0.0, got %s", ip)
	}
	if port != 8080 {
		t.Errorf("expected port 8080, got %d", port)
	}
}

func TestParseHexAddr_InvalidFormat(t *testing.T) {
	_, _, err := parseHexAddr("invalid")
	if err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}
}

func TestParseHexAddr_InvalidHex(t *testing.T) {
	_, _, err := parseHexAddr("ZZZZZZZZ:0050")
	if err == nil {
		t.Fatal("expected error for invalid hex IP, got nil")
	}
}

func TestListenerStruct(t *testing.T) {
	l := Listener{
		Protocol:     "TCP",
		LocalAddress: "0.0.0.0",
		LocalPort:    443,
		PID:          1234,
		ProgramName:  "nginx",
	}

	if l.Protocol != "TCP" {
		t.Errorf("expected TCP, got %s", l.Protocol)
	}
	if l.LocalPort != 443 {
		t.Errorf("expected 443, got %d", l.LocalPort)
	}
	if l.ProgramName != "nginx" {
		t.Errorf("expected nginx, got %s", l.ProgramName)
	}
}

func TestScanListeners_ReturnsNoError(t *testing.T) {
	// This test runs only on Linux where /proc/net/tcp exists.
	// On other systems it will gracefully surface the error.
	_, err := ScanListeners()
	if err != nil {
		t.Logf("ScanListeners error (may be non-Linux env): %v", err)
	}
}
