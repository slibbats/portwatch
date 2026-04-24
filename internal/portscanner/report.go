package portscanner

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"
	"time"
)

// ReportOptions configures how a port scan report is rendered.
type ReportOptions struct {
	Output      io.Writer
	ShowProcess bool
	ShowProtocol bool
}

// DefaultReportOptions returns sensible defaults for report generation.
func DefaultReportOptions() ReportOptions {
	return ReportOptions{
		Output:      os.Stdout,
		ShowProcess: true,
		ShowProtocol: true,
	}
}

// PrintReport writes a formatted table of listeners to the configured output.
func PrintReport(listeners []Listener, opts ReportOptions) {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	sorted := make([]Listener, len(listeners))
	copy(sorted, listeners)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Port != sorted[j].Port {
			return sorted[i].Port < sorted[j].Port
		}
		return sorted[i].Protocol < sorted[j].Protocol
	})

	w := tabwriter.NewWriter(opts.Output, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "# Port Report — %s\n", time.Now().Format(time.RFC3339))

	header := "PORT\tADDRESS"
	if opts.ShowProtocol {
		header += "\tPROTOCOL"
	}
	if opts.ShowProcess {
		header += "\tPROCESS"
	}
	fmt.Fprintln(w, header)

	for _, l := range sorted {
		line := fmt.Sprintf("%d\t%s", l.Port, l.Address)
		if opts.ShowProtocol {
			line += fmt.Sprintf("\t%s", l.Protocol)
		}
		if opts.ShowProcess {
			proc := l.Process
			if proc == "" {
				proc = "unknown"
			}
			line += fmt.Sprintf("\t%s", proc)
		}
		fmt.Fprintln(w, line)
	}
	w.Flush()
}
