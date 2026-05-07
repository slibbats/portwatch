package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/user/portwatch/internal/portscanner"
)

func runCorrelation(args []string) {
	fs := flag.NewFlagSet("correlation", flag.ExitOnError)
	dataDir := fs.String("data-dir", "/var/lib/portwatch", "directory for correlation data")
	fs.Parse(args)

	subArgs := fs.Args()
	if len(subArgs) == 0 {
		fmt.Fprintln(os.Stderr, "usage: portwatch correlation <set|get|remove|list> [port/proto] [incident-id]")
		os.Exit(1)
	}

	store := portscanner.NewCorrelationStore(*dataDir)
	_ = store.Load()

	switch subArgs[0] {
	case "set":
		if len(subArgs) < 3 {
			fmt.Fprintln(os.Stderr, "usage: portwatch correlation set <port/proto> <incident-id>")
			os.Exit(1)
		}
		l, err := parseCorrelationPortProto(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port/proto: %v\n", err)
			os.Exit(1)
		}
		store.Set(l, subArgs[2])
		if err := store.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "save failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("correlated port %d/%s with %s\n", l.Port, l.Protocol, subArgs[2])

	case "get":
		if len(subArgs) < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch correlation get <port/proto>")
			os.Exit(1)
		}
		l, err := parseCorrelationPortProto(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port/proto: %v\n", err)
			os.Exit(1)
		}
		val, ok := store.Get(l)
		if !ok {
			fmt.Printf("no correlation for %d/%s\n", l.Port, l.Protocol)
			return
		}
		fmt.Printf("%d/%s -> %s\n", l.Port, l.Protocol, val)

	case "remove":
		if len(subArgs) < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch correlation remove <port/proto>")
			os.Exit(1)
		}
		l, err := parseCorrelationPortProto(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port/proto: %v\n", err)
			os.Exit(1)
		}
		store.Remove(l)
		if err := store.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "save failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("removed correlation for %d/%s\n", l.Port, l.Protocol)

	case "list":
		all := store.All()
		if len(all) == 0 {
			fmt.Println("no correlations recorded")
			return
		}
		for _, e := range all {
			val, _ := store.Get(e)
			fmt.Printf("%d/%-5s %s\n", e.Port, e.Protocol, val)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", subArgs[0])
		os.Exit(1)
	}
}

func parseCorrelationPortProto(s string) (portscanner.Listener, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return portscanner.Listener{}, fmt.Errorf("expected port/proto, got %q", s)
	}
	port, err := strconv.Atoi(parts[0])
	if err != nil {
		return portscanner.Listener{}, fmt.Errorf("invalid port %q: %w", parts[0], err)
	}
	proto := strings.TrimSpace(parts[1])
	if proto == "" {
		return portscanner.Listener{}, fmt.Errorf("protocol must not be empty")
	}
	return portscanner.Listener{Port: port, Protocol: proto}, nil
}
