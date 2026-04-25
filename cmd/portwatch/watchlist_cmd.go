package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/user/portwatch/internal/portscanner"
)

func runWatchlist(args []string) error {
	fs := flag.NewFlagSet("watchlist", flag.ContinueOnError)
	file := fs.String("file", "watchlist.json", "path to watchlist JSON file")
	add := fs.Bool("add", false, "add a new entry to the watchlist")
	port := fs.Int("port", 0, "port number (used with --add)")
	proto := fs.String("proto", "tcp", "protocol: tcp or udp (used with --add)")
	process := fs.String("process", "", "process name (used with --add)")
	note := fs.String("note", "", "optional note (used with --add)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *add {
		if *port == 0 {
			return fmt.Errorf("watchlist: --port is required when using --add")
		}
		var w *portscanner.Watchlist
		loaded, err := portscanner.LoadWatchlist(*file)
		if err != nil {
			w = portscanner.NewWatchlist()
		} else {
			w = loaded
		}
		w.Add(*port, *proto, *process, *note)
		if err := portscanner.SaveWatchlist(*file, w); err != nil {
			return fmt.Errorf("watchlist: save: %w", err)
		}
		fmt.Fprintf(os.Stdout, "added %d/%s to watchlist\n", *port, *proto)
		return nil
	}

	// Default: list entries
	w, err := portscanner.LoadWatchlist(*file)
	if err != nil {
		return fmt.Errorf("watchlist: load: %w", err)
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PORT\tPROTOCOL\tPROCESS\tNOTE\tADDED")
	for _, e := range w.Entries {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n",
			e.Port, e.Protocol, e.Process, e.Note, e.AddedAt)
	}
	return tw.Flush()
}
