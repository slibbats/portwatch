package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

func runSuppression(args []string) error {
	fs := flag.NewFlagSet("suppression", flag.ContinueOnError)
	dataDir := fs.String("data-dir", "/var/lib/portwatch", "directory for persistent data")
	duration := fs.Duration("duration", time.Hour, "how long to suppress alerts (e.g. 2h30m)")
	reason := fs.String("reason", "", "reason for suppression")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: suppression <add|remove|list|prune> [port/proto]")
	}

	cmd := fs.Arg(0)

	store, err := portscanner.NewSuppressionStore(*dataDir)
	if err != nil {
		return fmt.Errorf("load suppression store: %w", err)
	}

	switch cmd {
	case "add":
		if fs.NArg() < 2 {
			return fmt.Errorf("add requires port/proto argument")
		}
		port, proto, err := parseSuppressionPortProto(fs.Arg(1))
		if err != nil {
			return err
		}
		until := time.Now().Add(*duration)
		if err := store.Set(port, proto, *reason, until); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "suppressed %d/%s until %s\n", port, proto, until.Format(time.RFC3339))

	case "remove":
		if fs.NArg() < 2 {
			return fmt.Errorf("remove requires port/proto argument")
		}
		port, proto, err := parseSuppressionPortProto(fs.Arg(1))
		if err != nil {
			return err
		}
		if err := store.Remove(port, proto); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "removed suppression for %d/%s\n", port, proto)

	case "list":
		entries := store.All()
		if len(entries) == 0 {
			fmt.Fprintln(os.Stdout, "no suppressions configured")
			return nil
		}
		fmt.Fprintf(os.Stdout, "%-10s %-6s %-30s %s\n", "PORT", "PROTO", "UNTIL", "REASON")
		for _, e := range entries {
			active := "active"
			if !e.IsActive() {
				active = "expired"
			}
			fmt.Fprintf(os.Stdout, "%-10d %-6s %-30s %s [%s]\n",
				e.Port, e.Proto, e.Until.Format(time.RFC3339), e.Reason, active)
		}

	case "prune":
		if err := store.PruneExpired(); err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, "pruned expired suppressions")

	default:
		return fmt.Errorf("unknown subcommand: %s", cmd)
	}
	return nil
}

func parseSuppressionPortProto(s string) (int, string, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid format %q: expected port/proto", s)
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
