package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/user/portwatch/internal/portscanner"
)

func runNote(args []string) error {
	fs := flag.NewFlagSet("note", flag.ContinueOnError)
	dir := fs.String("dir", "/var/lib/portwatch", "directory for note storage")
	proto := fs.String("proto", "tcp", "protocol (tcp/udp)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	path := filepath.Join(*dir, "notes.json")
	store, err := portscanner.NewNoteStore(path)
	if err != nil {
		return fmt.Errorf("open note store: %w", err)
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		return fmt.Errorf("usage: note <add|remove|list> [port] [text]")
	}

	switch strings.ToLower(remaining[0]) {
	case "add":
		if len(remaining) < 3 {
			return fmt.Errorf("usage: note add <port> <text>")
		}
		port, err := strconv.Atoi(remaining[1])
		if err != nil {
			return fmt.Errorf("invalid port %q: %w", remaining[1], err)
		}
		text := strings.Join(remaining[2:], " ")
		store.Set(port, *proto, text)
		if err := store.Save(); err != nil {
			return fmt.Errorf("save: %w", err)
		}
		fmt.Fprintf(os.Stdout, "note added for %s/%d\n", *proto, port)

	case "remove":
		if len(remaining) < 2 {
			return fmt.Errorf("usage: note remove <port>")
		}
		port, err := strconv.Atoi(remaining[1])
		if err != nil {
			return fmt.Errorf("invalid port %q: %w", remaining[1], err)
		}
		if !store.Remove(port, *proto) {
			return fmt.Errorf("no note found for %s/%d", *proto, port)
		}
		if err := store.Save(); err != nil {
			return fmt.Errorf("save: %w", err)
		}
		fmt.Fprintf(os.Stdout, "note removed for %s/%d\n", *proto, port)

	case "list":
		notes := store.All()
		if len(notes) == 0 {
			fmt.Fprintln(os.Stdout, "no notes stored")
			return nil
		}
		for _, n := range notes {
			fmt.Fprintf(os.Stdout, "%s/%d\t%s\t(updated: %s)\n",
				n.Protocol, n.Port, n.Text, n.UpdatedAt.Format("2006-01-02 15:04:05"))
		}

	default:
		return fmt.Errorf("unknown subcommand %q; use add, remove, or list", remaining[0])
	}

	return nil
}
