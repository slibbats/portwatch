package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/user/portwatch/internal/portscanner"
)

func runReport(args []string) error {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)

	snapshotDir := fs.String("snapshot-dir", "/var/lib/portwatch/snapshots", "Directory containing snapshots")
	format := fs.String("format", "table", "Output format: table or compact")
	sortBy := fs.String("sort", "port", "Sort field: port or process")
	latest := fs.Bool("latest", true, "Use the latest snapshot")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("report: flag parse error: %w", err)
	}

	store := portscanner.NewHistoryStore(*snapshotDir)

	var snap *portscanner.Snapshot
	var err error

	if *latest {
		snap, err = store.Latest()
		if err != nil {
			return fmt.Errorf("report: could not load latest snapshot: %w", err)
		}
	} else {
		snaps, err := store.All()
		if err != nil || len(snaps) == 0 {
			return fmt.Errorf("report: no snapshots found in %s", *snapshotDir)
		}
		snap = snaps[len(snaps)-1]
	}

	opts := portscanner.DefaultReportOptions()
	opts.Output = os.Stdout
	opts.Format = *format
	opts.SortBy = *sortBy

	portscanner.PrintReport(snap.Listeners, opts)
	return nil
}
