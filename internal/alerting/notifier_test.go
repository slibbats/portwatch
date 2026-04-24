package alerting

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

func makeTestListener(port uint16, pid int) portscanner.Listener {
	return portscanner.Listener{
		Address: "0.0.0.0",
		Port:    port,
		PID:     pid,
	}
}

func TestNotifier_Notify_WritesFormattedAlert(t *testing.T) {
	var buf bytes.Buffer
	n := NewNotifier(&buf)

	a := Alert{
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Severity:  SeverityWarning,
		Message:   "unexpected new listener detected",
		Listener:  makeTestListener(8080, 1234),
	}

	if err := n.Notify(a); err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "WARNING") {
		t.Errorf("expected WARNING in output, got: %s", out)
	}
	if !strings.Contains(out, "8080") {
		t.Errorf("expected port 8080 in output, got: %s", out)
	}
	if !strings.Contains(out, "1234") {
		t.Errorf("expected pid 1234 in output, got: %s", out)
	}
}

func TestNotifier_NotifyNew_EmitsOneLinePerListener(t *testing.T) {
	var buf bytes.Buffer
	n := NewNotifier(&buf)

	listeners := []portscanner.Listener{
		makeTestListener(9000, 100),
		makeTestListener(9001, 101),
	}

	if err := n.NotifyNew(listeners); err != nil {
		t.Fatalf("NotifyNew returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines of output, got %d: %s", len(lines), buf.String())
	}
}

func TestNewNotifier_NilOutputDefaultsToStdout(t *testing.T) {
	n := NewNotifier(nil)
	if n.out == nil {
		t.Error("expected non-nil writer when nil passed to NewNotifier")
	}
}

func TestNotifier_NotifyNew_EmptySlice(t *testing.T) {
	var buf bytes.Buffer
	n := NewNotifier(&buf)

	if err := n.NotifyNew([]portscanner.Listener{}); err != nil {
		t.Fatalf("NotifyNew returned error on empty slice: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty listeners, got: %s", buf.String())
	}
}
