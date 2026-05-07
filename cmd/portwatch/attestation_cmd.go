package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/example/portwatch/internal/portscanner"
)

func runAttestation(args []string) {
	fs := flag.NewFlagSet("attestation", flag.ExitOnError)
	dataDir := fs.String("data-dir", "/var/lib/portwatch", "directory for portwatch data")
	attestBy := fs.String("by", "", "name of the person attesting")
	reason := fs.String("reason", "", "reason for attestation")
	expiry := fs.String("expires", "", "optional expiry duration (e.g. 24h, 7d)")
	_ = fs.Parse(args)

	subArgs := fs.Args()
	if len(subArgs) == 0 {
		fmt.Fprintln(os.Stderr, "usage: portwatch attestation <add|remove|list> [port/proto]")
		os.Exit(1)
	}

	store := portscanner.NewAttestationStore(*dataDir)
	cmd := subArgs[0]

	switch cmd {
	case "list":
		all, err := store.All()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if len(all) == 0 {
			fmt.Println("no attestations recorded")
			return
		}
		for _, a := range all {
			expStr := "never"
			if a.ExpiresAt != nil {
				expStr = a.ExpiresAt.Format(time.RFC3339)
			}
			fmt.Printf("%d/%s\tattested-by=%s\treason=%s\texpires=%s\n",
				a.Port, a.Proto, a.AttestedBy, a.Reason, expStr)
		}

	case "add":
		if len(subArgs) < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch attestation add <port/proto>")
			os.Exit(1)
		}
		port, proto, err := parseAttestationPortProto(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port/proto: %v\n", err)
			os.Exit(1)
		}
		a := portscanner.Attestation{
			Port:       port,
			Proto:      proto,
			AttestedBy: *attestBy,
			Reason:     *reason,
			AttestedAt: time.Now().UTC(),
		}
		if *expiry != "" {
			d, err := time.ParseDuration(*expiry)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid expiry duration: %v\n", err)
				os.Exit(1)
			}
			t := time.Now().Add(d)
			a.ExpiresAt = &t
		}
		if err := store.Set(a); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("attested %d/%s\n", port, proto)

	case "remove":
		if len(subArgs) < 2 {
			fmt.Fprintln(os.Stderr, "usage: portwatch attestation remove <port/proto>")
			os.Exit(1)
		}
		port, proto, err := parseAttestationPortProto(subArgs[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid port/proto: %v\n", err)
			os.Exit(1)
		}
		if err := store.Remove(port, proto); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("removed attestation for %d/%s\n", port, proto)

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", cmd)
		os.Exit(1)
	}
}

func parseAttestationPortProto(s string) (int, string, error) {
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
