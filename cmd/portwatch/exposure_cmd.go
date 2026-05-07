package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/user/portwatch/internal/portscanner"
)

func runExposure(args []string) {
	fs := flag.NewFlagSet("exposure", flag.ExitOnError)
	storeDir := fs.String("store", "/var/lib/portwatch", "path to data store directory")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: portwatch exposure <set|get|remove|list> [port/proto] [level]")
		fmt.Fprintln(os.Stderr, "levels: public, internal, private, unknown")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	store, err := portscanner.NewExposureStore(*storeDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "exposure: failed to open store: %v\n", err)
		os.Exit(1)
	}

	switch fs.Arg(0) {
	case "set":
		if fs.NArg() < 3 {
			fmt.Fprintln(os.Stderr, "usage: portwatch exposure set <port/proto> <level>")
			os.Exit(1)
		}
		port, proto, err := parseExposurePortProto(fs.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "exposure: %v\n", err)
			os.Exit(1)
		}
		level, err := portscanner.ParseExposureLevel(fs.Arg(2))
		if err != nil {
			fmt.Fprintf(os.Stderr, "exposure: %v\n", err)
			os.Exit(1)
		}
		if err := store.Set(port, proto, level); err != nil {
			fmt.Fprintf(os.Stderr, "exposure: set: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("exposure set: %d/%s = %s\n", port, proto, level)

	case "get":
		if fs.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch exposure get <port/proto>")
			os.Exit(1)
		}
		port, proto, err := parseExposurePortProto(fs.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "exposure: %v\n", err)
			os.Exit(1)
		}
		level, ok := store.Get(port, proto)
		if !ok {
			fmt.Printf("no exposure set for %d/%s\n", port, proto)
			return
		}
		fmt.Printf("%d/%s: %s\n", port, proto, level)

	case "remove":
		if fs.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch exposure remove <port/proto>")
			os.Exit(1)
		}
		port, proto, err := parseExposurePortProto(fs.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "exposure: %v\n", err)
			os.Exit(1)
		}
		if err := store.Remove(port, proto); err != nil {
			fmt.Fprintf(os.Stderr, "exposure: remove: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("removed exposure for %d/%s\n", port, proto)

	case "list":
		entries := store.All()
		if len(entries) == 0 {
			fmt.Println("no exposure entries")
			return
		}
		for _, e := range entries {
			fmt.Printf("%-8d %-6s %s\n", e.Port, e.Proto, e.Exposure)
		}

	default:
		fs.Usage()
		os.Exit(1)
	}
}

func parseExposurePortProto(s string) (int, string, error) {
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
