package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/user/portwatch/internal/portscanner"
)

func runCompliance(args []string) error {
	fs := flag.NewFlagSet("compliance", flag.ContinueOnError)
	storeDir := fs.String("dir", ".portwatch/compliance", "directory for compliance data")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: portwatch compliance <set|get|remove|list> [port/proto] [status] [policy] [reason]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	cs, err := portscanner.NewComplianceStore(*storeDir)
	if err != nil {
		return fmt.Errorf("compliance: %w", err)
	}

	positional := fs.Args()
	if len(positional) == 0 {
		fs.Usage()
		return fmt.Errorf("compliance: subcommand required")
	}

	switch positional[0] {
	case "set":
		if len(positional) < 4 {
			return fmt.Errorf("compliance set: usage: set <port/proto> <status> <policy> [reason]")
		}
		port, proto, err := parseCompliancePortProto(positional[1])
		if err != nil {
			return err
		}
		status, err := portscanner.ParseComplianceStatus(positional[2])
		if err != nil {
			return err
		}
		reason := ""
		if len(positional) >= 5 {
			reason = positional[4]
		}
		if err := cs.Set(port, proto, status, positional[3], reason); err != nil {
			return err
		}
		fmt.Printf("compliance: set %s/%s -> %s (%s)\n", positional[1], proto, status, positional[3])

	case "get":
		if len(positional) < 2 {
			return fmt.Errorf("compliance get: usage: get <port/proto>")
		}
		port, proto, err := parseCompliancePortProto(positional[1])
		if err != nil {
			return err
		}
		e, ok := cs.Get(port, proto)
		if !ok {
			return fmt.Errorf("compliance: no entry for %s", positional[1])
		}
		fmt.Printf("port=%d proto=%s status=%s policy=%s reason=%s\n",
			e.Port, e.Proto, e.Status, e.Policy, e.Reason)

	case "remove":
		if len(positional) < 2 {
			return fmt.Errorf("compliance remove: usage: remove <port/proto>")
		}
		port, proto, err := parseCompliancePortProto(positional[1])
		if err != nil {
			return err
		}
		if err := cs.Remove(port, proto); err != nil {
			return err
		}
		fmt.Printf("compliance: removed %s\n", positional[1])

	case "list":
		all := cs.All()
		if len(all) == 0 {
			fmt.Println("no compliance entries")
			return nil
		}
		for _, e := range all {
			fmt.Printf("%-6d %-5s %-8s %-12s %s\n", e.Port, e.Proto, e.Status, e.Policy, e.Reason)
		}

	default:
		return fmt.Errorf("compliance: unknown subcommand %q", positional[0])
	}
	return nil
}

func parseCompliancePortProto(s string) (int, string, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("compliance: invalid port/proto %q (expected port/proto)", s)
	}
	port, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("compliance: invalid port %q: %w", parts[0], err)
	}
	if parts[1] == "" {
		return 0, "", fmt.Errorf("compliance: proto must not be empty")
	}
	return port, parts[1], nil
}
