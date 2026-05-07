package portscanner

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func makeBaselineDiffListener(port int, proto, addr, process string) Listener {
	return Listener{
		Port:     port,
		Protocol: proto,
		Address:  addr,
		Process:  process,
	}
}

func TestPrintBaselineDiff_NoChanges(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultBaselineDiffOptions()
	opts.Output = &buf

	result := BaselineDiffResult{
		New:        []Listener{},
		Removed:    []Listener{},
		CapturedAt: time.Now(),
	}
	PrintBaselineDiff(result, opts)

	if !strings.Contains(buf.String(), "No changes detected") {
		t.Errorf("expected no-changes message, got: %s", buf.String())
	}
}

func TestPrintBaselineDiff_ShowsNewListeners(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultBaselineDiffOptions()
	opts.Output = &buf

	result := BaselineDiffResult{
		New:        []Listener{makeBaselineDiffListener(8080, "tcp", "0.0.0.0", "nginx")},
		Removed:    []Listener{},
		CapturedAt: time.Now(),
	}
	PrintBaselineDiff(result, opts)

	out := buf.String()
	if !strings.Contains(out, "NEW LISTENERS") {
		t.Errorf("expected NEW LISTENERS header, got: %s", out)
	}
	if !strings.Contains(out, "8080") {
		t.Errorf("expected port 8080 in output, got: %s", out)
	}
	if !strings.Contains(out, "nginx") {
		t.Errorf("expected process nginx in output, got: %s", out)
	}
}

func TestPrintBaselineDiff_ShowsRemovedListeners(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultBaselineDiffOptions()
	opts.Output = &buf
	opts.ShowRemoved = true

	result := BaselineDiffResult{
		New:        []Listener{},
		Removed:    []Listener{makeBaselineDiffListener(22, "tcp", "0.0.0.0", "sshd")},
		CapturedAt: time.Now(),
	}
	PrintBaselineDiff(result, opts)

	out := buf.String()
	if !strings.Contains(out, "REMOVED LISTENERS") {
		t.Errorf("expected REMOVED LISTENERS header, got: %s", out)
	}
	if !strings.Contains(out, "22") {
		t.Errorf("expected port 22 in output, got: %s", out)
	}
}

func TestPrintBaselineDiff_HidesRemovedWhenDisabled(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultBaselineDiffOptions()
	opts.Output = &buf
	opts.ShowRemoved = false

	result := BaselineDiffResult{
		New:        []Listener{makeBaselineDiffListener(9090, "tcp", "127.0.0.1", "app")},
		Removed:    []Listener{makeBaselineDiffListener(22, "tcp", "0.0.0.0", "sshd")},
		CapturedAt: time.Now(),
	}
	PrintBaselineDiff(result, opts)

	out := buf.String()
	if strings.Contains(out, "REMOVED LISTENERS") {
		t.Errorf("expected REMOVED LISTENERS to be hidden, got: %s", out)
	}
	if !strings.Contains(out, "9090") {
		t.Errorf("expected port 9090 in output, got: %s", out)
	}
}

func TestPrintBaselineDiff_UnknownProcessFallback(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultBaselineDiffOptions()
	opts.Output = &buf

	result := BaselineDiffResult{
		New:        []Listener{makeBaselineDiffListener(3000, "tcp", "0.0.0.0", "")},
		CapturedAt: time.Now(),
	}
	PrintBaselineDiff(result, opts)

	if !strings.Contains(buf.String(), "unknown") {
		t.Errorf("expected 'unknown' process fallback, got: %s", buf.String())
	}
}

func TestPrintBaselineDiff_NilOutputDefaultsToStdout(t *testing.T) {
	opts := DefaultBaselineDiffOptions()
	opts.Output = nil

	// Should not panic
	PrintBaselineDiff(BaselineDiffResult{CapturedAt: time.Now()}, opts)
}
