package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/user/portwatch/internal/portscanner"
)

func runEscalation(args []string) {
	fs := flag.NewFlagSet("escalation", flag.ExitOnError)
	dataDir := fs.String("data-dir", "/var/lib/portwatch", "directory for portwatch data")
	_ = fs.Parse(args)

	subArgs := fs.Args()
	if len(subArgs) == 0 {
		fmt.Fprintln(os.Stderr, "usage: portwatch escalation <add|remove|list> [args]")
		os.Exit(1)
	}

	store, err := portscanner.NewEscalationStore(*dataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening escalation store: %v\n", err)
		os.Exit(1)
	}

	switch subArgs[0] {
	case "add":
		if len(subArgs) < 5 {
			fmt.Fprintln(os.Stderr, "usage: portwatch escalation add <port/proto> <contact> <channel> <min-level>")
			os.Exit(1)
		}
		port, proto, err := portscanner.ParseEscalationPortProto(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port/proto: %v\n", err)
			os.Exit(1)
		}
		if err := store.Set(port, proto, subArgs[2], subArgs[3], subArgs[4]); err != nil {
			fmt.Fprintf(os.Stderr, "error saving escalation policy: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("escalation policy set for %d/%s\n", port, proto)

	case "remove":
		if len(subArgs) < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch escalation remove <port/proto>")
			os.Exit(1)
		}
		port, proto, err := portscanner.ParseEscalationPortProto(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port/proto: %v\n", err)
			os.Exit(1)
		}
		if err := store.Remove(port, proto); err != nil {
			fmt.Fprintf(os.Stderr, "error removing escalation policy: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("escalation policy removed for %d/%s\n", port, proto)

	case "list":
		all := store.All()
		if len(all) == 0 {
			fmt.Println("no escalation policies defined")
			return
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PORT\tPROTO\tCONTACT\tCHANNEL\tMIN LEVEL")
		for _, p := range all {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", p.Port, p.Proto, p.Contact, p.Channel, p.MinLevel)
		}
		_ = w.Flush()

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", subArgs[0])
		os.Exit(1)
	}
}
