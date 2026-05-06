package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/user/portwatch/internal/portscanner"
)

func runRemediation(args []string) {
	fs := flag.NewFlagSet("remediation", flag.ExitOnError)
	dataDir := fs.String("data-dir", "/var/lib/portwatch", "directory for remediation data")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: portwatch remediation <set|get|remove|list> [port/proto] [action] [--script=...] [--run-as=...]")
		fs.PrintDefaults()
	}
	fs.Parse(args)

	if fs.NArg() < 1 {
		fs.Usage()
		os.Exit(1)
	}

	store, err := portscanner.NewRemediationStore(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	switch fs.Arg(0) {
	case "set":
		if fs.NArg() < 3 {
			fmt.Fprintln(os.Stderr, "usage: remediation set <port/proto> <action> [script] [run-as]")
			os.Exit(1)
		}
		port, proto, err := parseRemediationPortProto(fs.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		r := portscanner.RemediationAction{Port: port, Proto: proto, Action: fs.Arg(2)}
		if fs.NArg() > 3 {
			r.Script = fs.Arg(3)
		}
		if fs.NArg() > 4 {
			r.RunAs = fs.Arg(4)
		}
		if err := store.Set(r); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("remediation set for %d/%s\n", port, proto)
	case "get":
		if fs.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "usage: remediation get <port/proto>")
			os.Exit(1)
		}
		port, proto, err := parseRemediationPortProto(fs.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		r, ok := store.Get(port, proto)
		if !ok {
			fmt.Fprintf(os.Stderr, "no remediation for %d/%s\n", port, proto)
			os.Exit(1)
		}
		fmt.Printf("port=%d proto=%s action=%s script=%s run_as=%s\n", r.Port, r.Proto, r.Action, r.Script, r.RunAs)
	case "remove":
		if fs.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "usage: remediation remove <port/proto>")
			os.Exit(1)
		}
		port, proto, err := parseRemediationPortProto(fs.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if err := store.Remove(port, proto); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("removed remediation for %d/%s\n", port, proto)
	case "list":
		for _, r := range store.All() {
			fmt.Printf("%d/%s\taction=%s\tscript=%s\trun_as=%s\n", r.Port, r.Proto, r.Action, r.Script, r.RunAs)
		}
	default:
		fs.Usage()
		os.Exit(1)
	}
}

func parseRemediationPortProto(s string) (int, string, error) {
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
