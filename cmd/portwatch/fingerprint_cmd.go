package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/user/portwatch/internal/portscanner"
)

func runFingerprint(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("fingerprint", flag.ContinueOnError)
	fs.SetOutput(out)

	var (
		dataDir  = fs.String("data-dir", "/var/lib/portwatch", "directory to store fingerprint")
		diffMode = fs.Bool("diff", false, "diff current listeners against saved fingerprint")
		save     = fs.Bool("save", false, "save current listeners as fingerprint baseline")
	)

	if err := fs.Parse(args); err != nil {
		return err
	}

	listeners, err := portscanner.ScanListeners()
	if err != nil {
		return fmt.Errorf("fingerprint: scan: %w", err)
	}

	current := portscanner.NewFingerprint(listeners)

	if *save {
		if err := portscanner.SaveFingerprint(current, *dataDir); err != nil {
			return err
		}
		fmt.Fprintf(out, "Fingerprint saved: %d entries\n", len(current.Entries))
		return nil
	}

	if *diffMode {
		baseline, err := portscanner.LoadFingerprint(*dataDir)
		if err != nil {
			return fmt.Errorf("fingerprint: load baseline: %w", err)
		}
		added, removed := portscanner.DiffFingerprint(baseline, current)
		if len(added) == 0 && len(removed) == 0 {
			fmt.Fprintln(out, "No changes from fingerprint baseline.")
			return nil
		}
		for _, e := range added {
			fmt.Fprintf(out, "+ %d/%s\t%s\t%s\n", e.Port, e.Protocol, e.Address, e.Process)
		}
		for _, e := range removed {
			fmt.Fprintf(out, "- %d/%s\t%s\t%s\n", e.Port, e.Protocol, e.Address, e.Process)
		}
		return nil
	}

	// Default: print current fingerprint
	for _, e := range current.Entries {
		fmt.Fprintf(out, "%d/%s\t%s\t%s\n", e.Port, e.Protocol, e.Address, e.Process)
	}
	return nil
}

func init() {
	_ = os.Stderr // ensure os is used
}
