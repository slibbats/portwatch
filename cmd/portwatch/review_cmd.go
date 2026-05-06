package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/user/portwatch/internal/portscanner"
)

func runReview(args []string) {
	fs := flag.NewFlagSet("review", flag.ExitOnError)
	dataDir := fs.String("data-dir", ".portwatch", "directory for review data")
	reviewerFlag := fs.String("reviewer", "", "reviewer name")
	noteFlag := fs.String("note", "", "optional review note")
	_ = fs.Parse(args)

	subArgs := fs.Args()
	if len(subArgs) < 1 {
		fmt.Fprintln(os.Stderr, "usage: portwatch review <add|remove|list> [port/proto] [status]")
		os.Exit(1)
	}

	store, err := portscanner.NewReviewStore(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "review: %v\n", err)
		os.Exit(1)
	}

	switch subArgs[0] {
	case "add":
		if len(subArgs) < 3 {
			fmt.Fprintln(os.Stderr, "usage: portwatch review add <port/proto> <approved|rejected|pending>")
			os.Exit(1)
		}
		port, proto, err := parseReviewPortProto(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "review: %v\n", err)
			os.Exit(1)
		}
		status := portscanner.ReviewStatus(subArgs[2])
		if err := store.Set(port, proto, status, *reviewerFlag, *noteFlag); err != nil {
			fmt.Fprintf(os.Stderr, "review: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("review: set %d/%s => %s\n", port, proto, status)

	case "remove":
		if len(subArgs) < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch review remove <port/proto>")
			os.Exit(1)
		}
		port, proto, err := parseReviewPortProto(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "review: %v\n", err)
			os.Exit(1)
		}
		if err := store.Remove(port, proto); err != nil {
			fmt.Fprintf(os.Stderr, "review: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("review: removed %d/%s\n", port, proto)

	case "list":
		for _, e := range store.All() {
			fmt.Printf("%d/%s\tstatus=%-10s\treviewer=%s\tnote=%s\n", e.Port, e.Proto, e.Status, e.Reviewer, e.Note)
		}

	default:
		fmt.Fprintf(os.Stderr, "review: unknown subcommand %q\n", subArgs[0])
		os.Exit(1)
	}
}

func parseReviewPortProto(s string) (int, string, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid format %q, expected port/proto", s)
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
