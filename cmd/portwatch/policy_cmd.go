package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/iamcalledned/portwatch/internal/portscanner"
)

func runPolicy(args []string) {
	fs := flag.NewFlagSet("policy", flag.ExitOnError)
	dir := fs.String("dir", ".portwatch/policies", "policy store directory")
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: portwatch policy <add|remove|list> [port/proto] [allow|deny] [reason]")
		os.Exit(1)
	}

	store, err := portscanner.NewPolicyStore(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "policy: %v\n", err)
		os.Exit(1)
	}

	switch fs.Arg(0) {
	case "add":
		if fs.NArg() < 3 {
			fmt.Fprintln(os.Stderr, "usage: portwatch policy add <port/proto> <allow|deny> [reason]")
			os.Exit(1)
		}
		port, proto, err := parsePolicyPortProto(fs.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "policy add: %v\n", err)
			os.Exit(1)
		}
		action := fs.Arg(2)
		if action != "allow" && action != "deny" {
			fmt.Fprintln(os.Stderr, "action must be 'allow' or 'deny'")
			os.Exit(1)
		}
		reason := ""
		if fs.NArg() >= 4 {
			reason = strings.Join(fs.Args()[3:], " ")
		}
		if err := store.Set(port, proto, action, reason); err != nil {
			fmt.Fprintf(os.Stderr, "policy add: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("policy set: %d/%s -> %s\n", port, proto, action)

	case "remove":
		if fs.NArg() < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch policy remove <port/proto>")
			os.Exit(1)
		}
		port, proto, err := parsePolicyPortProto(fs.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "policy remove: %v\n", err)
			os.Exit(1)
		}
		if err := store.Remove(port, proto); err != nil {
			fmt.Fprintf(os.Stderr, "policy remove: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("policy removed: %d/%s\n", port, proto)

	case "list":
		policies := store.All()
		if len(policies) == 0 {
			fmt.Println("no policies defined")
			return
		}
		fmt.Printf("%-8s %-6s %-6s %s\n", "PORT", "PROTO", "ACTION", "REASON")
		for _, p := range policies {
			fmt.Printf("%-8d %-6s %-6s %s\n", p.Port, p.Proto, p.Action, p.Reason)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", fs.Arg(0))
		os.Exit(1)
	}
}

func parsePolicyPortProto(s string) (int, string, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid format %q, expected port/proto", s)
	}
	port, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("invalid port %q: %w", parts[0], err)
	}
	if parts[1] == "" {
		return 0, "", fmt.Errorf("proto must not be empty")
	}
	return port, parts[1], nil
}
