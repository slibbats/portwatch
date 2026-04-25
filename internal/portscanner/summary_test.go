package portscanner_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/portscanner"
)

func makeSummaryListener(addr, proto string, port uint16, pid int, process string) portscanner.Listener {
	return portscanner.Listener{
		Address:  addr,
		Protocol: proto,
		Port:     port,
		PID:      pid,
		Process:  process,
	}
}

func TestSummarize_CountsByProtocol(t *testing.T) {
	listeners := []portscanner.Listener{
		makeSummaryListener("0.0.0.0", "tcp", 80, 100, "nginx"),
		makeSummaryListener("0.0.0.0", "tcp", 443, 100, "nginx"),
		makeSummaryListener("0.0.0.0", "udp", 53, 200, "dnsmasq"),
	}

	summary := portscanner.Summarize(listeners)

	if summary.TotalCount != 3 {
		t.Errorf("expected TotalCount=3, got %d", summary.TotalCount)
	}
	if summary.TCPCount != 2 {
		t.Errorf("expected TCPCount=2, got %d", summary.TCPCount)
	}
	if summary.UDPCount != 1 {
		t.Errorf("expected UDPCount=1, got %d", summary.UDPCount)
	}
}

func TestSummarize_UniqueProcesses(t *testing.T) {
	listeners := []portscanner.Listener{
		makeSummaryListener("0.0.0.0", "tcp", 80, 100, "nginx"),
		makeSummaryListener("0.0.0.0", "tcp", 443, 100, "nginx"),
		makeSummaryListener("0.0.0.0", "udp", 53, 200, "dnsmasq"),
	}

	summary := portscanner.Summarize(listeners)

	if summary.UniqueProcesses != 2 {
		t.Errorf("expected UniqueProcesses=2, got %d", summary.UniqueProcesses)
	}
}

func TestSummarize_EmptyListeners(t *testing.T) {
	summary := portscanner.Summarize([]portscanner.Listener{})

	if summary.TotalCount != 0 {
		t.Errorf("expected TotalCount=0, got %d", summary.TotalCount)
	}
	if summary.UniqueProcesses != 0 {
		t.Errorf("expected UniqueProcesses=0, got %d", summary.UniqueProcesses)
	}
}

func TestPrintSummary_ContainsExpectedFields(t *testing.T) {
	listeners := []portscanner.Listener{
		makeSummaryListener("0.0.0.0", "tcp", 8080, 300, "myapp"),
		makeSummaryListener("127.0.0.1", "udp", 5353, 400, "avahi"),
	}

	var buf bytes.Buffer
	portscanner.PrintSummary(listeners, &buf)
	out := buf.String()

	for _, want := range []string{"Total", "TCP", "UDP", "Process"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestPrintSummary_NilOutputDefaultsToStdout(t *testing.T) {
	// Should not panic when output is nil
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PrintSummary panicked with nil output: %v", r)
		}
	}()
	portscanner.PrintSummary([]portscanner.Listener{}, nil)
}
