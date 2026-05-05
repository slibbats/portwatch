package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/example/portwatch/internal/portscanner"
)

func runRateLimit(args []string) error {
	fs := flag.NewFlagSet("ratelimit", flag.ContinueOnError)
	dataDir := fs.String("data-dir", "/var/lib/portwatch", "directory for persistent data")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: portwatch ratelimit <add|remove|list> [port/proto] [max-per-hour]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	store, err := portscanner.NewRateLimitStore(*dataDir)
	if err != nil {
		return fmt.Errorf("failed to open rate limit store: %w", err)
	}

	positional := fs.Args()
	if len(positional) == 0 {
		fs.Usage()
		return fmt.Errorf("subcommand required: add, remove, or list")
	}

	switch positional[0] {
	case "add":
		if len(positional) < 3 {
			return fmt.Errorf("usage: ratelimit add <port/proto> <max-per-hour>")
		}
		port, proto, err := parseRateLimitPortProto(positional[1])
		if err != nil {
			return err
		}
		max, err := strconv.Atoi(positional[2])
		if err != nil || max <= 0 {
			return fmt.Errorf("max-per-hour must be a positive integer")
		}
		if err := store.Set(port, proto, max); err != nil {
			return err
		}
		fmt.Printf("rate limit set: %d/%s → %d/hr\n", port, proto, max)

	case "remove":
		if len(positional) < 2 {
			return fmt.Errorf("usage: ratelimit remove <port/proto>")
		}
		port, proto, err := parseRateLimitPortProto(positional[1])
		if err != nil {
			return err
		}
		if err := store.Remove(port, proto); err != nil {
			return err
		}
		fmt.Printf("rate limit removed: %d/%s\n", port, proto)

	case "list":
		entries := store.All()
		if len(entries) == 0 {
			fmt.Println("no rate limits configured")
			return nil
		}
		fmt.Printf("%-10s %-8s %s\n", "PORT", "PROTO", "MAX/HR")
		for _, e := range entries {
			fmt.Printf("%-10d %-8s %d\n", e.Port, e.Proto, e.MaxPerHour)
		}

	default:
		return fmt.Errorf("unknown subcommand: %s", positional[0])
	}
	return nil
}

func parseRateLimitPortProto(s string) (int, string, error) {
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
