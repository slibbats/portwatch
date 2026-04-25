package portscanner

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"
)

// TrendReportOptions controls output of the trend report.
type TrendReportOptions struct {
	Output    io.Writer
	MinCount  int
	Since     time.Duration
}

// DefaultTrendReportOptions returns sensible defaults.
func DefaultTrendReportOptions() TrendReportOptions {
	return TrendReportOptions{
		Output:   os.Stdout,
		MinCount: 1,
		Since:    24 * time.Hour,
	}
}

// PrintTrendReport renders a TrendResult table to the configured output.
func PrintTrendReport(results []TrendResult, opts TrendReportOptions) {
	out := opts.Output
	if out == nil {
		out = os.Stdout
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PORT\tPROTOCOL\tAPPEARANCES\tFIRST SEEN\tLAST SEEN\tSTABLE")
	fmt.Fprintln(w, "----\t--------\t-----------\t----------\t---------\t------")

	for _, r := range results {
		if r.AppearanceCount < opts.MinCount {
			continue
		}
		stable := "no"
		if r.IsStable {
			stable = "yes"
		}
		fmt.Fprintf(w, "%d\t%s\t%d\t%s\t%s\t%s\n",
			r.Port,
			r.Protocol,
			r.AppearanceCount,
			r.FirstSeen.Format(time.RFC3339),
			r.LastSeen.Format(time.RFC3339),
			stable,
		)
	}
	w.Flush()
}
