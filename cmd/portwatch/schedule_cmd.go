package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/user/portwatch/internal/portscanner"
)

func runSchedule(args []string) {
	fs := flag.NewFlagSet("schedule", flag.ExitOnError)
	dataDir := fs.String("data-dir", "/var/lib/portwatch", "directory for schedule data")
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: portwatch schedule <add|remove|list> [port/proto] [cron] [label]")
		os.Exit(1)
	}

	store, err := portscanner.NewScheduleStore(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading schedule store: %v\n", err)
		os.Exit(1)
	}

	switch fs.Arg(0) {
	case "add":
		if fs.NArg() < 3 {
			fmt.Fprintln(os.Stderr, "usage: portwatch schedule add <port/proto> <cron> [label]")
			os.Exit(1)
		}
		port, proto, err := parseSchedulePortProto(fs.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port/proto: %v\n", err)
			os.Exit(1)
		}
		cron := fs.Arg(2)
		label := ""
		if fs.NArg() >= 4 {
			label = strings.Join(fs.Args()[3:], " ")
		}
		if err := store.Set(port, proto, cron, label); err != nil {
			fmt.Fprintf(os.Stderr, "error saving schedule: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("schedule added for %d/%s\n", port, proto)

	case "remove":
		if fs.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch schedule remove <port/proto>")
			os.Exit(1)
		}
		port, proto, err := parseSchedulePortProto(fs.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port/proto: %v\n", err)
			os.Exit(1)
		}
		if err := store.Remove(port, proto); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("schedule removed for %d/%s\n", port, proto)

	case "list":
		entries := store.All()
		if len(entries) == 0 {
			fmt.Println("no schedules defined")
			return
		}
		fmt.Printf("%-10s %-6s %-20s %s\n", "PORT", "PROTO", "CRON", "LABEL")
		for _, e := range entries {
			fmt.Printf("%-10d %-6s %-20s %s\n", e.Port, e.Proto, e.Cron, e.Label)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", fs.Arg(0))
		os.Exit(1)
	}
}

func parseSchedulePortProto(s string) (int, string, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("expected port/proto format, got %q", s)
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
