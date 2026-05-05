package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/user/portwatch/internal/portscanner"
)

func runChangelog(args []string) {
	fs := flag.NewFlagSet("changelog", flag.ExitOnError)
	snapshotDir := fs.String("snapshot-dir", "/var/lib/portwatch/snapshots", "directory for snapshot data")
	port := fs.Int("port", 0, "filter by port number (0 = all)")
	proto := fs.String("proto", "tcp", "protocol to filter by when -port is set")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: portwatch changelog [flags]")
		fmt.Fprintln(os.Stderr, "\nShow the changelog of port listener additions and removals.")
		fmt.Fprintln(os.Stderr, "\nFlags:")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	store := portscanner.NewChangelogStore(*snapshotDir)

	var entries []portscanner.ChangelogEntry
	var err error

	if *port != 0 {
		entries, err = store.FilterByPort(*port, *proto)
	} else {
		entries, err = store.All()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading changelog: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("No changelog entries found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tEVENT\tPORT\tPROTO\tADDRESS\tPROCESS")
	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			e.Timestamp.Format("2006-01-02T15:04:05Z"),
			e.Event,
			e.Port,
			e.Proto,
			e.Address,
			e.Process,
		)
	}
	_ = w.Flush()
}
