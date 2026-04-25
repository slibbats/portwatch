package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

func runTrend(args []string) {
	fs := flag.NewFlagSet("trend", flag.ContinueOnError)
	snapshotDir := fs.String("snapshot-dir", "/var/lib/portwatch/snapshots", "Directory containing snapshots")
	minCount := fs.Int("min-count", 1, "Minimum appearance count to include in report")
	since := fs.Duration("since", 24*time.Hour, "Only consider snapshots within this duration")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return
		}
		fmt.Fprintf(os.Stderr, "trend: %v\n", err)
		os.Exit(1)
	}

	store, err := portscanner.NewHistoryStore(*snapshotDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "trend: failed to open history store: %v\n", err)
		os.Exit(1)
	}

	snapshots, err := store.All()
	if err != nil {
		fmt.Fprintf(os.Stderr, "trend: failed to load snapshots: %v\n", err)
		os.Exit(1)
	}

	cutoff := time.Now().Add(-*since)
	var filtered []*portscanner.Snapshot
	for _, s := range snapshots {
		if s.Timestamp.After(cutoff) {
			filtered = append(filtered, s)
		}
	}

	results := portscanner.AnalyzeTrends(filtered)
	opts := portscanner.DefaultTrendReportOptions()
	opts.MinCount = *minCount
	portscanner.PrintTrendReport(results, opts)
}
