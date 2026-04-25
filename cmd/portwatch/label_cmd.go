package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/user/portwatch/internal/portscanner"
)

func runLabel(args []string) error {
	fs := flag.NewFlagSet("label", flag.ContinueOnError)
	dir := fs.String("state-dir", "/var/lib/portwatch", "directory for state files")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: portwatch label <add|remove|list> [port] [proto] [label]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	positional := fs.Args()
	if len(positional) == 0 {
		fs.Usage()
		return fmt.Errorf("subcommand required: add, remove, or list")
	}

	store, err := portscanner.NewLabelStore(*dir)
	if err != nil {
		return fmt.Errorf("label: open store: %w", err)
	}

	switch positional[0] {
	case "add":
		if len(positional) < 4 {
			return fmt.Errorf("usage: label add <port> <proto> <label>")
		}
		port, err := strconv.ParseUint(positional[1], 10, 16)
		if err != nil {
			return fmt.Errorf("invalid port %q: %w", positional[1], err)
		}
		store.Set(uint16(port), positional[2], positional[3])
		if err := store.Save(); err != nil {
			return fmt.Errorf("label: save: %w", err)
		}
		fmt.Printf("labeled %s/%s as %q\n", positional[1], positional[2], positional[3])

	case "remove":
		if len(positional) < 3 {
			return fmt.Errorf("usage: label remove <port> <proto>")
		}
		port, err := strconv.ParseUint(positional[1], 10, 16)
		if err != nil {
			return fmt.Errorf("invalid port %q: %w", positional[1], err)
		}
		store.Remove(uint16(port), positional[2])
		if err := store.Save(); err != nil {
			return fmt.Errorf("label: save: %w", err)
		}
		fmt.Printf("removed label for %s/%s\n", positional[1], positional[2])

	case "list":
		all := store.All()
		if len(all) == 0 {
			fmt.Println("no labels defined")
			return nil
		}
		for key, label := range all {
			fmt.Printf("  %-20s %s\n", key, label)
		}

	default:
		return fmt.Errorf("unknown subcommand %q", positional[0])
	}
	return nil
}
