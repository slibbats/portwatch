package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/user/portwatch/internal/portscanner"
)

func runComment(args []string) {
	fs := flag.NewFlagSet("comment", flag.ExitOnError)
	dataDir := fs.String("data-dir", "/var/lib/portwatch", "directory for comment storage")
	_ = fs.Parse(args)

	subArgs := fs.Args()
	if len(subArgs) < 1 {
		fmt.Fprintln(os.Stderr, "usage: portwatch comment <set|get|remove|list> [port/proto] [text]")
		os.Exit(1)
	}

	store, err := portscanner.NewCommentStore(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading comment store: %v\n", err)
		os.Exit(1)
	}

	switch subArgs[0] {
	case "set":
		if len(subArgs) < 3 {
			fmt.Fprintln(os.Stderr, "usage: portwatch comment set <port/proto> <text>")
			os.Exit(1)
		}
		port, proto, err := parsePortProto(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port/proto: %v\n", err)
			os.Exit(1)
		}
		text := strings.Join(subArgs[2:], " ")
		if err := store.Set(port, proto, text); err != nil {
			fmt.Fprintf(os.Stderr, "error saving comment: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("comment set for %d/%s\n", port, proto)

	case "get":
		if len(subArgs) < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch comment get <port/proto>")
			os.Exit(1)
		}
		port, proto, err := parsePortProto(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port/proto: %v\n", err)
			os.Exit(1)
		}
		c, ok := store.Get(port, proto)
		if !ok {
			fmt.Printf("no comment for %d/%s\n", port, proto)
			return
		}
		fmt.Printf("%d/%s: %s (updated %s)\n", port, proto, c.Text, c.UpdatedAt.Format("2006-01-02 15:04:05"))

	case "remove":
		if len(subArgs) < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch comment remove <port/proto>")
			os.Exit(1)
		}
		port, proto, err := parsePortProto(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port/proto: %v\n", err)
			os.Exit(1)
		}
		if err := store.Remove(port, proto); err != nil {
			fmt.Fprintf(os.Stderr, "error removing comment: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("comment removed for %d/%s\n", port, proto)

	case "list":
		all := store.All()
		if len(all) == 0 {
			fmt.Println("no comments stored")
			return
		}
		for key, c := range all {
			fmt.Printf("%-12s %s\n", key, c.Text)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", subArgs[0])
		os.Exit(1)
	}
}

func parsePortProto(s string) (int, string, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("expected port/proto format, got %q", s)
	}
	port, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("invalid port %q: %w", parts[0], err)
	}
	return port, parts[1], nil
}
