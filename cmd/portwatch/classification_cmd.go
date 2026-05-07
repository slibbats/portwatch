package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/user/portwatch/internal/portscanner"
)

func runClassification(args []string) {
	fs := flag.NewFlagSet("classification", flag.ExitOnError)
	dataDir := fs.String("data-dir", "/var/lib/portwatch", "data directory")
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: portwatch classification <add|remove|list> [port/proto] [level] [rationale]")
		os.Exit(1)
	}

	store, err := portscanner.NewClassificationStore(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "classification: %v\n", err)
		os.Exit(1)
	}

	switch fs.Arg(0) {
	case "add":
		if fs.NArg() < 3 {
			fmt.Fprintln(os.Stderr, "usage: portwatch classification add <port/proto> <level> [rationale]")
			os.Exit(1)
		}
		port, proto, err := parseClassificationPortProto(fs.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "classification add: %v\n", err)
			os.Exit(1)
		}
		level, err := portscanner.ParseClassificationLevel(fs.Arg(2))
		if err != nil {
			fmt.Fprintf(os.Stderr, "classification add: %v\n", err)
			os.Exit(1)
		}
		rationale := ""
		if fs.NArg() >= 4 {
			rationale = strings.Join(fs.Args()[3:], " ")
		}
		if err := store.Set(port, proto, level, rationale); err != nil {
			fmt.Fprintf(os.Stderr, "classification add: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("classification set: %d/%s = %s\n", port, proto, level)

	case "remove":
		if fs.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch classification remove <port/proto>")
			os.Exit(1)
		}
		port, proto, err := parseClassificationPortProto(fs.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "classification remove: %v\n", err)
			os.Exit(1)
		}
		if err := store.Remove(port, proto); err != nil {
			fmt.Fprintf(os.Stderr, "classification remove: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("classification removed: %d/%s\n", port, proto)

	case "list":
		entries := store.All()
		if len(entries) == 0 {
			fmt.Println("no classification entries")
			return
		}
		fmt.Printf("%-10s %-6s %-14s %s\n", "PORT", "PROTO", "LEVEL", "RATIONALE")
		for _, e := range entries {
			fmt.Printf("%-10d %-6s %-14s %s\n", e.Port, e.Proto, e.Level, e.Rationale)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", fs.Arg(0))
		os.Exit(1)
	}
}

func parseClassificationPortProto(s string) (int, string, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("expected port/proto, got %q", s)
	}
	port, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("invalid port %q: %w", parts[0], err)
	}
	proto := strings.TrimSpace(parts[1])
	if proto == "" {
		return 0, "", fmt.Errorf("proto must not be empty")
	}
	return port, proto, nil
}
