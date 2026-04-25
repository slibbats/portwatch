package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// runExport handles the `portwatch export` sub-command.
// Usage: portwatch export [-format json|csv] [-output <file>]
func runExport(args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	format := fs.String("format", "json", "output format: json or csv")
	outputPath := fs.String("output", "", "write output to file instead of stdout")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("export: %w", err)
	}

	listeners, err := portscanner.ScanListeners()
	if err != nil {
		return fmt.Errorf("export: scan failed: %w", err)
	}

	opts := portscanner.DefaultExportOptions()
	opts.Format = portscanner.ExportFormat(*format)
	opts.Timestamp = time.Now()

	if *outputPath != "" {
		f, err := os.Create(*outputPath)
		if err != nil {
			return fmt.Errorf("export: cannot create output file: %w", err)
		}
		defer f.Close()
		opts.Output = f
	} else {
		opts.Output = os.Stdout
	}

	if err := portscanner.Export(listeners, opts); err != nil {
		return fmt.Errorf("export: %w", err)
	}
	return nil
}
