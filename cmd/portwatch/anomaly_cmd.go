package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

func runAnomaly(args []string) error {
	fs := flag.NewFlagSet("anomaly", flag.ContinueOnError)

	snapshotDir := fs.String("snapshot-dir", "/var/lib/portwatch/snapshots", "Directory containing snapshots")
	window := fs.Duration("window", 24*time.Hour, "Time window to analyse (e.g. 1h, 24h)")
	minFlap := fs.Int("min-flap", 3, "Minimum flap count to flag a port as flapping")

	if err := fs.Parse(args); err != nil {
		return err
	}

	store := portscanner.NewHistoryStore(*snapshotDir)

	snapshots, err := store.All()
	if err != nil {
		return fmt.Errorf("loading snapshots: %w", err)
	}

	if len(snapshots) == 0 {
		fmt.Fprintln(os.Stdout, "No snapshots found.")
		return nil
	}

	cutoff := time.Now().Add(-*window)
	var filtered []*portscanner.Snapshot
	for _, s := range snapshots {
		if s.Timestamp.After(cutoff) {
			filtered = append(filtered, s)
		}
	}

	if len(filtered) == 0 {
		fmt.Fprintf(os.Stdout, "No snapshots within the last %s.\n", *window)
		return nil
	}

	opts := portscanner.AnomalyOptions{
		MinFlapCount: *minFlap,
	}

	anomalies := portscanner.DetectAnomalies(filtered, opts)

	if len(anomalies) == 0 {
		fmt.Fprintln(os.Stdout, "No anomalies detected.")
		return nil
	}

	fmt.Fprintf(os.Stdout, "Detected %d anomaly(ies):\n\n", len(anomalies))
	for _, a := range anomalies {
		fmt.Fprintf(os.Stdout, "  [%s] port %d/%s — %s\n", a.Kind, a.Port, a.Protocol, a.Detail)
	}

	return nil
}
