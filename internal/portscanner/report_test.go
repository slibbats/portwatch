package portscanner

import (
	"bytes"
	"strings"
	"testing"
)

func makeReportListener(port uint16, addr, proto, process string) Listener {
	return Listener{
		Port:     port,
		Address:  addr,
		Protocol: proto,
		Process:  process,
	}
}

func TestPrintReport_ContainsPortAndAddress(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultReportOptions()
	opts.Output = &buf

	listeners := []Listener{
		makeReportListener(8080, "0.0.0.0", "tcp", "nginx"),
	}
	PrintReport(listeners, opts)

	out := buf.String()
	if !strings.Contains(out, "8080") {
		t.Errorf("expected port 8080 in output, got: %s", out)
	}
	if !strings.Contains(out, "0.0.0.0") {
		t.Errorf("expected address in output, got: %s", out)
	}
}

func TestPrintReport_SortedByPort(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultReportOptions()
	opts.Output = &buf

	listeners := []Listener{
		makeReportListener(9090, "0.0.0.0", "tcp", "app"),
		makeReportListener(80, "0.0.0.0", "tcp", "nginx"),
		makeReportListener(443, "0.0.0.0", "tcp", "nginx"),
	}
	PrintReport(listeners, opts)

	out := buf.String()
	idx80 := strings.Index(out, "80")
	idx443 := strings.Index(out, "443")
	idx9090 := strings.Index(out, "9090")

	if idx80 > idx443 || idx443 > idx9090 {
		t.Errorf("expected ports sorted ascending, got output:\n%s", out)
	}
}

func TestPrintReport_UnknownProcessFallback(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultReportOptions()
	opts.Output = &buf

	listeners := []Listener{
		makeReportListener(3000, "127.0.0.1", "tcp", ""),
	}
	PrintReport(listeners, opts)

	if !strings.Contains(buf.String(), "unknown") {
		t.Errorf("expected 'unknown' for empty process name")
	}
}

func TestPrintReport_NilOutputDefaultsToStdout(t *testing.T) {
	// Should not panic when Output is nil
	opts := ReportOptions{Output: nil, ShowProcess: true, ShowProtocol: true}
	listeners := []Listener{
		makeReportListener(22, "0.0.0.0", "tcp", "sshd"),
	}
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintReport panicked with nil output: %v", r)
		}
	}()
	PrintReport(listeners, opts)
}

func TestDefaultReportOptions_Defaults(t *testing.T) {
	opts := DefaultReportOptions()
	if opts.Output == nil {
		t.Error("expected non-nil default output")
	}
	if !opts.ShowProcess {
		t.Error("expected ShowProcess to be true by default")
	}
	if !opts.ShowProtocol {
		t.Error("expected ShowProtocol to be true by default")
	}
}
