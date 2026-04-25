package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch subcmd {
	case "export":
		err = runExport(args)
	case "summary":
		err = runSummary(args)
	case "report":
		err = runReport(args)
	case "help", "--help", "-h":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "portwatch: unknown command %q\n", subcmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "portwatch %s: %v\n", subcmd, err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `portwatch — monitor port usage and alert on unexpected listeners

Usage:
  portwatch <command> [flags]

Commands:
  export     Export a snapshot to JSON or CSV
  summary    Print a summary of the latest snapshot
  report     Print a detailed port listener report
  help       Show this help message

Run 'portwatch <command> --help' for command-specific flags.`)
}
