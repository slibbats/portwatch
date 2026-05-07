package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/example/portwatch/internal/portscanner"
)

func runBaselineDiff(args []string) {
	fs := flag.NewFlagSet("baseline-diff", flag.ExitOnError)
	baselineFile := fs.String("baseline", "baseline.json", "Path to baseline file")
	showRemoved := fs.Bool("show-removed", true, "Show ports removed since baseline")
	excludeLoopback := fs.Bool("exclude-loopback", false, "Exclude loopback addresses")
	excludeIPv6 := fs.Bool("exclude-ipv6", false, "Exclude IPv6 listeners")

	_ = fs.Parse(args)

	baseline, err := portscanner.LoadBaseline(*baselineFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading baseline: %v\n", err)
		os.Exit(1)
	}

	filterOpts := portscanner.DefaultFilterOptions()
	filterOpts.ExcludeLoopback = *excludeLoopback
	filterOpts.ExcludeIPv6 = *excludeIPv6

	current, err := portscanner.ScanListeners()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error scanning listeners: %v\n", err)
		os.Exit(1)
	}
	current = filterOpts.Apply(current)

	newListeners, removedListeners := baseline.Diff(current)

	result := portscanner.BaselineDiffResult{
		New:        newListeners,
		Removed:    removedListeners,
		CapturedAt: baseline.CapturedAt,
	}

	diffOpts := portscanner.DefaultBaselineDiffOptions()
	diffOpts.ShowRemoved = *showRemoved

	portscanner.PrintBaselineDiff(result, diffOpts)
}
