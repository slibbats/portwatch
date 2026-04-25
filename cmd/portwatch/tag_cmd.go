package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/example/portwatch/internal/portscanner"
)

func runTag(args []string) error {
	fs := flag.NewFlagSet("tag", flag.ContinueOnError)
	tagFile := fs.String("file", "tags.json", "path to tag store file")
	remove := fs.Bool("remove", false, "remove the tag for the given port/protocol")
	note := fs.String("note", "", "optional note for the tag")
	list := fs.Bool("list", false, "list all tags")

	if err := fs.Parse(args); err != nil {
		return err
	}

	ts, err := portscanner.LoadTagStore(*tagFile)
	if err != nil {
		return fmt.Errorf("load tag store: %w", err)
	}

	if *list {
		tags := ts.Sorted()
		if len(tags) == 0 {
			fmt.Println("No tags defined.")
			return nil
		}
		fmt.Printf("%-8s %-8s %-20s %s\n", "PORT", "PROTO", "LABEL", "NOTE")
		for _, t := range tags {
			fmt.Printf("%-8d %-8s %-20s %s\n", t.Port, t.Protocol, t.Label, t.Note)
		}
		return nil
	}

	remaining := fs.Args()
	if len(remaining) < 2 {
		return fmt.Errorf("usage: tag [flags] <port/proto> <label>\n  e.g. tag 8080/tcp my-service")
	}

	parts := strings.SplitN(remaining[0], "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid port/proto format %q, expected e.g. 8080/tcp", remaining[0])
	}
	port, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid port %q: %w", parts[0], err)
	}
	protocol := strings.ToLower(parts[1])

	if *remove {
		if !ts.Remove(port, protocol) {
			fmt.Fprintf(os.Stderr, "warning: no tag found for %d/%s\n", port, protocol)
		}
	} else {
		label := remaining[1]
		ts.Add(portscanner.Tag{Port: port, Protocol: protocol, Label: label, Note: *note})
	}

	if err := portscanner.SaveTagStore(*tagFile, ts); err != nil {
		return fmt.Errorf("save tag store: %w", err)
	}
	fmt.Printf("Tag store updated: %s\n", *tagFile)
	return nil
}
