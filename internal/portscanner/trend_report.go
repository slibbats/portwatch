package portscanner

import (
	"fmt"
	"io"
	"os"
	"sort"
)

// TrendReportOptions configures how a trend report is rendered.
type TrendReportOptions struct {
	// MinFrequency filters out entries below this ratio (0.0–1.0).
	MinFrequency float64
	// Output is the writer for report output; defaults to os.Stdout.
	Output io.Writer
}

// DefaultTrendReportOptions returns sensible defaults.
func DefaultTrendReportOptions() TrendReportOptions {
	return TrendReportOptions{
		MinFrequency: 0.0,
		Output:       os.Stdout,
	}
}

// PrintTrendReport writes a formatted trend report from a slice of snapshots.
func PrintTrendReport(snapshots []*Snapshot, opts TrendReportOptions) {
	out := opts.Output
	if out == nil {
		out = os.Stdout
	}

	entries := AnalyzeTrends(snapshots)

	// Filter by minimum frequency.
	filtered := entries[:0]
	for _, e := range entries {
		if e.FrequencyRatio() >= opts.MinFrequency {
			filtered = append(filtered, e)
		}
	}

	// Sort descending by frequency, then by port for stable output.
	sort.Slice(filtered, func(i, j int) bool {
		ri, rj := filtered[i].FrequencyRatio(), filtered[j].FrequencyRatio()
		if ri != rj {
			return ri > rj
		}
		return filtered[i].Port < filtered[j].Port
	})

	fmt.Fprintf(out, "%-6s %-20s %6s  %s\n", "PROTO", "ADDRESS", "PORT", "FREQUENCY")
	fmt.Fprintf(out, "%-6s %-20s %6s  %s\n", "-----", "-------", "----", "---------")
	for _, e := range filtered {
		fmt.Fprintf(out, "%-6s %-20s %6d  %d/%d (%.0f%%)\n",
			e.Proto, e.Addr, e.Port,
			e.SeenIn, e.TotalSnapshots,
			e.FrequencyRatio()*100,
		)
	}

	if len(filtered) == 0 {
		fmt.Fprintln(out, "no trend data available")
	}
}
