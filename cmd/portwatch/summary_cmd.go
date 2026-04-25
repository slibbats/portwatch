package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/user/portwatch/internal/portscanner"
)

func runSummary(args []string) error {
	fs := flag.NewFlagSet("summary", flag.ContinueOnError)

	snapshotDir := fs.String("snapshot-dir", "/var/lib/portwatch/snapshots", "Directory containing snapshots")
	latest := fs.Bool("latest", false, "Summarize only the latest snapshot")
	liveScan := fs.Bool("live", false, "Run a live scan instead of loading a snapshot")

	if err := fs.Parse(args); err != nil {
		return err
	}

	var listeners []portscanner.Listener

	switch {
	case *liveScan:
		var err error
		listeners, err = portscanner.ScanListeners()
		if err != nil {
			return fmt.Errorf("live scan failed: %w", err)
		}

	case *latest:
		store := portscanner.NewHistoryStore(*snapshotDir)
		snap, err := store.Latest()
		if err != nil {
			return fmt.Errorf("failed to load latest snapshot: %w", err)
		}
		listeners = snap.Listeners

	default:
		var err error
		listeners, err = portscanner.ScanListeners()
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}
	}

	portscanner.PrintSummary(listeners, os.Stdout)
	return nil
}
