package portscanner

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"
	"time"
)

// BaselineDiffResult holds the result of comparing current listeners to a baseline.
type BaselineDiffResult struct {
	New     []Listener
	Removed []Listener
	CapturedAt time.Time
}

// DefaultBaselineDiffOptions returns sensible defaults for baseline diff output.
func DefaultBaselineDiffOptions() BaselineDiffOptions {
	return BaselineDiffOptions{
		Output:      os.Stdout,
		ShowRemoved: true,
	}
}

// BaselineDiffOptions controls how a baseline diff is printed.
type BaselineDiffOptions struct {
	Output      io.Writer
	ShowRemoved bool
}

// PrintBaselineDiff writes a human-readable diff between the baseline and current listeners.
func PrintBaselineDiff(result BaselineDiffResult, opts BaselineDiffOptions) {
	out := opts.Output
	if out == nil {
		out = os.Stdout
	}

	fmt.Fprintf(out, "Baseline captured at: %s\n\n", result.CapturedAt.Format(time.RFC3339))

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	if len(result.New) == 0 && (!opts.ShowRemoved || len(result.Removed) == 0) {
		fmt.Fprintln(out, "No changes detected from baseline.")
		return
	}

	if len(result.New) > 0 {
		fmt.Fprintln(w, "[+] NEW LISTENERS")
		fmt.Fprintln(w, "PROTO\tADDRESS\tPORT\tPROCESS")
		sort.Slice(result.New, func(i, j int) bool {
			return result.New[i].Port < result.New[j].Port
		})
		for _, l := range result.New {
			proc := l.Process
			if proc == "" {
				proc = "unknown"
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", l.Protocol, l.Address, l.Port, proc)
		}
		_ = w.Flush()
	}

	if opts.ShowRemoved && len(result.Removed) > 0 {
		fmt.Fprintln(w, "\n[-] REMOVED LISTENERS")
		fmt.Fprintln(w, "PROTO\tADDRESS\tPORT\tPROCESS")
		sort.Slice(result.Removed, func(i, j int) bool {
			return result.Removed[i].Port < result.Removed[j].Port
		})
		for _, l := range result.Removed {
			proc := l.Process
			if proc == "" {
				proc = "unknown"
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", l.Protocol, l.Address, l.Port, proc)
		}
		_ = w.Flush()
	}
}
