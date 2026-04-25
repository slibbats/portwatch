package portscanner

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

// ExportFormat defines the output format for exporting listener data.
type ExportFormat string

const (
	ExportJSON ExportFormat = "json"
	ExportCSV  ExportFormat = "csv"
)

// ExportOptions controls how listeners are exported.
type ExportOptions struct {
	Format    ExportFormat
	Output    io.Writer
	Timestamp time.Time
}

// DefaultExportOptions returns ExportOptions with sensible defaults.
func DefaultExportOptions() ExportOptions {
	return ExportOptions{
		Format:    ExportJSON,
		Output:    os.Stdout,
		Timestamp: time.Now(),
	}
}

// exportRecord is the serialisable representation of a single listener.
type exportRecord struct {
	Timestamp string `json:"timestamp" csv:"timestamp"`
	Protocol  string `json:"protocol"  csv:"protocol"`
	Address   string `json:"address"   csv:"address"`
	Port      int    `json:"port"      csv:"port"`
	PID       int    `json:"pid"       csv:"pid"`
	Process   string `json:"process"   csv:"process"`
}

// Export writes listeners to the configured output in the requested format.
func Export(listeners []Listener, opts ExportOptions) error {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	ts := opts.Timestamp.UTC().Format(time.RFC3339)

	records := make([]exportRecord, len(listeners))
	for i, l := range listeners {
		records[i] = exportRecord{
			Timestamp: ts,
			Protocol:  l.Protocol,
			Address:   l.Address,
			Port:      l.Port,
			PID:       l.PID,
			Process:   l.Process,
		}
	}

	switch opts.Format {
	case ExportCSV:
		return exportCSV(records, opts.Output)
	case ExportJSON:
		return exportJSON(records, opts.Output)
	default:
		return fmt.Errorf("unsupported export format: %q", opts.Format)
	}
}

func exportJSON(records []exportRecord, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}

func exportCSV(records []exportRecord, w io.Writer) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"timestamp", "protocol", "address", "port", "pid", "process"}); err != nil {
		return err
	}
	for _, r := range records {
		row := []string{r.Timestamp, r.Protocol, r.Address, strconv.Itoa(r.Port), strconv.Itoa(r.PID), r.Process}
		if err := cw.Write(row); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
