package portscanner

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func makeExportListener(proto, addr string, port, pid int, process string) Listener {
	return Listener{
		Protocol: proto,
		Address:  addr,
		Port:     port,
		PID:      pid,
		Process:  process,
	}
}

func TestExport_JSON_ContainsExpectedFields(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultExportOptions()
	opts.Output = &buf
	opts.Format = ExportJSON
	opts.Timestamp = time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	listeners := []Listener{
		makeExportListener("tcp", "0.0.0.0", 8080, 42, "nginx"),
	}

	if err := Export(listeners, opts); err != nil {
		t.Fatalf("Export returned error: %v", err)
	}

	var records []exportRecord
	if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	r := records[0]
	if r.Protocol != "tcp" {
		t.Errorf("expected protocol tcp, got %q", r.Protocol)
	}
	if r.Port != 8080 {
		t.Errorf("expected port 8080, got %d", r.Port)
	}
	if r.Process != "nginx" {
		t.Errorf("expected process nginx, got %q", r.Process)
	}
	if r.Timestamp != "2024-01-15T10:00:00Z" {
		t.Errorf("unexpected timestamp: %q", r.Timestamp)
	}
}

func TestExport_CSV_ContainsHeader(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultExportOptions()
	opts.Output = &buf
	opts.Format = ExportCSV

	listeners := []Listener{
		makeExportListener("udp", "127.0.0.1", 53, 10, "dnsmasq"),
	}

	if err := Export(listeners, opts); err != nil {
		t.Fatalf("Export returned error: %v", err)
	}

	r := csv.NewReader(&buf)
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}
	if len(rows) < 2 {
		t.Fatalf("expected header + 1 data row, got %d rows", len(rows))
	}
	if rows[0][0] != "timestamp" {
		t.Errorf("expected first header column 'timestamp', got %q", rows[0][0])
	}
	if rows[1][1] != "udp" {
		t.Errorf("expected protocol udp in data row, got %q", rows[1][1])
	}
}

func TestExport_NilOutput_DefaultsToStdout(t *testing.T) {
	opts := DefaultExportOptions()
	opts.Output = nil
	// Should not panic — falls back to os.Stdout
	_ = Export([]Listener{}, opts)
}

func TestExport_UnsupportedFormat_ReturnsError(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultExportOptions()
	opts.Output = &buf
	opts.Format = ExportFormat("xml")

	err := Export([]Listener{}, opts)
	if err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported export format") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestExport_JSON_EmptyListeners(t *testing.T) {
	var buf bytes.Buffer
	opts := DefaultExportOptions()
	opts.Output = &buf
	opts.Format = ExportJSON

	if err := Export([]Listener{}, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var records []exportRecord
	if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}
