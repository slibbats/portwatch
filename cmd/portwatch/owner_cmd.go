package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/user/portwatch/internal/portscanner"
)

func runOwner(args []string) {
	fs := flag.NewFlagSet("owner", flag.ExitOnError)
	dataDir := fs.String("data-dir", "/var/lib/portwatch", "directory for portwatch data")
	_ = fs.Parse(args)

	subArgs := fs.Args()
	if len(subArgs) == 0 {
		fmt.Fprintln(os.Stderr, "usage: portwatch owner <add|remove|list> [port/proto] [owner] [--team T] [--email E]")
		os.Exit(1)
	}

	store, err := portscanner.NewOwnerStore(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading owner store: %v\n", err)
		os.Exit(1)
	}

	switch subArgs[0] {
	case "add":
		if len(subArgs) < 3 {
			fmt.Fprintln(os.Stderr, "usage: portwatch owner add <port/proto> <owner> [team] [email]")
			os.Exit(1)
		}
		port, proto, err := parseOwnerPortProto(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port/proto: %v\n", err)
			os.Exit(1)
		}
		e := portscanner.OwnerEntry{Port: port, Proto: proto, Owner: subArgs[2]}
		if len(subArgs) > 3 {
			e.Team = subArgs[3]
		}
		if len(subArgs) > 4 {
			e.Email = subArgs[4]
		}
		if err := store.Set(e); err != nil {
			fmt.Fprintf(os.Stderr, "error saving owner: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("owner set for %d/%s\n", port, proto)

	case "remove":
		if len(subArgs) < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch owner remove <port/proto>")
			os.Exit(1)
		}
		port, proto, err := parseOwnerPortProto(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port/proto: %v\n", err)
			os.Exit(1)
		}
		if err := store.Remove(port, proto); err != nil {
			fmt.Fprintf(os.Stderr, "error removing owner: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("owner removed for %d/%s\n", port, proto)

	case "list":
		all := store.All()
		if len(all) == 0 {
			fmt.Println("no owners defined")
			return
		}
		fmt.Printf("%-8s %-6s %-20s %-15s %s\n", "PORT", "PROTO", "OWNER", "TEAM", "EMAIL")
		for _, e := range all {
			fmt.Printf("%-8d %-6s %-20s %-15s %s\n", e.Port, e.Proto, e.Owner, e.Team, e.Email)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", subArgs[0])
		os.Exit(1)
	}
}

func parseOwnerPortProto(s string) (int, string, error) {
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
