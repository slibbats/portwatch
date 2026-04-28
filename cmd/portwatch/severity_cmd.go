package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/example/portwatch/internal/portscanner"
)

func runSeverity(args []string) {
	fs := flag.NewFlagSet("severity", flag.ExitOnError)
	dataDir := fs.String("data-dir", "/var/lib/portwatch", "directory for portwatch data")
	_ = fs.Parse(args)

	subArgs := fs.Args()
	if len(subArgs) < 1 {
		fmt.Fprintln(os.Stderr, "usage: portwatch severity <add|remove|list> [port] [protocol] [level]")
		os.Exit(1)
	}

	path := filepath.Join(*dataDir, "severity.json")
	store, err := portscanner.NewSeverityStore(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading severity store: %v\n", err)
		os.Exit(1)
	}

	switch subArgs[0] {
	case "add":
		if len(subArgs) < 4 {
			fmt.Fprintln(os.Stderr, "usage: portwatch severity add <port> <protocol> <low|medium|high|critical>")
			os.Exit(1)
		}
		port, err := strconv.Atoi(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port: %v\n", err)
			os.Exit(1)
		}
		level := portscanner.SeverityLevel(subArgs[3])
		if err := store.Set(port, subArgs[2], level); err != nil {
			fmt.Fprintf(os.Stderr, "error saving severity: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("severity set: %s/%d → %s\n", subArgs[2], port, level)

	case "remove":
		if len(subArgs) < 3 {
			fmt.Fprintln(os.Stderr, "usage: portwatch severity remove <port> <protocol>")
			os.Exit(1)
		}
		port, err := strconv.Atoi(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port: %v\n", err)
			os.Exit(1)
		}
		if err := store.Remove(port, subArgs[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error removing severity: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("severity removed: %s/%d\n", subArgs[2], port)

	case "list":
		rules := store.All()
		if len(rules) == 0 {
			fmt.Println("no severity rules defined")
			return
		}
		fmt.Printf("%-10s %-10s %s\n", "PROTOCOL", "PORT", "SEVERITY")
		for _, r := range rules {
			fmt.Printf("%-10s %-10d %s\n", r.Protocol, r.Port, r.Level)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", subArgs[0])
		os.Exit(1)
	}
}
