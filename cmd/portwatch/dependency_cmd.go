package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/user/portwatch/internal/portscanner"
)

func runDependency(args []string) error {
	fs := flag.NewFlagSet("dependency", flag.ContinueOnError)
	dataDir := fs.String("data-dir", "/var/lib/portwatch", "directory for dependency data")
	if err := fs.Parse(args); err != nil {
		return err
	}

	subArgs := fs.Args()
	if len(subArgs) == 0 {
		fmt.Fprintln(os.Stderr, "usage: dependency <add|remove|list> [port/proto] [dep1/proto ...]")
		return fmt.Errorf("subcommand required")
	}

	store, err := portscanner.NewDependencyStore(*dataDir)
	if err != nil {
		return fmt.Errorf("load dependency store: %w", err)
	}

	switch subArgs[0] {
	case "add":
		if len(subArgs) < 3 {
			return fmt.Errorf("add requires port/proto and at least one dependency")
		}
		port, proto, err := parseDependencyPortProto(subArgs[1])
		if err != nil {
			return err
		}
		var refs []portscanner.DependencyRef
		for _, raw := range subArgs[2:] {
			dp, dproto, err := parseDependencyPortProto(raw)
			if err != nil {
				return fmt.Errorf("invalid dep %q: %w", raw, err)
			}
			refs = append(refs, portscanner.DependencyRef{Port: dp, Proto: dproto})
		}
		if err := store.Set(port, proto, refs); err != nil {
			return err
		}
		fmt.Printf("dependency set for %d/%s\n", port, proto)
	case "remove":
		if len(subArgs) < 2 {
			return fmt.Errorf("remove requires port/proto")
		}
		port, proto, err := parseDependencyPortProto(subArgs[1])
		if err != nil {
			return err
		}
		if err := store.Remove(port, proto); err != nil {
			return err
		}
		fmt.Printf("dependency removed for %d/%s\n", port, proto)
	case "list":
		for _, d := range store.All() {
			var deps []string
			for _, r := range d.DependsOn {
				deps = append(deps, fmt.Sprintf("%d/%s", r.Port, r.Proto))
			}
			fmt.Printf("%d/%s -> [%s]\n", d.Port, d.Proto, strings.Join(deps, ", "))
		}
	default:
		return fmt.Errorf("unknown subcommand: %s", subArgs[0])
	}
	return nil
}

func parseDependencyPortProto(s string) (int, string, error) {
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
