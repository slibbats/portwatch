package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/user/portwatch/internal/portscanner"
)

func runAlertRule(args []string) error {
	fs := flag.NewFlagSet("alert-rule", flag.ContinueOnError)
	dir := fs.String("data-dir", "/var/lib/portwatch", "directory for storing alert rules")
	if err := fs.Parse(args); err != nil {
		return err
	}

	subArgs := fs.Args()
	if len(subArgs) == 0 {
		return fmt.Errorf("usage: alert-rule <add|remove|list> [flags]")
	}

	store, err := portscanner.NewAlertRuleStore(*dir)
	if err != nil {
		return fmt.Errorf("failed to load alert rules: %w", err)
	}

	switch subArgs[0] {
	case "add":
		if len(subArgs) < 3 {
			return fmt.Errorf("usage: alert-rule add <port/proto> <severity> [message]")
		}
		port, proto, err := parseAlertPortProto(subArgs[1])
		if err != nil {
			return err
		}
		msg := ""
		if len(subArgs) >= 4 {
			msg = strings.Join(subArgs[3:], " ")
		}
		store.Set(portscanner.AlertRule{
			Port:     port,
			Protocol: proto,
			Severity: subArgs[2],
			Message:  msg,
		})
		if err := store.Save(); err != nil {
			return fmt.Errorf("failed to save: %w", err)
		}
		fmt.Fprintf(os.Stdout, "alert rule added for %s/%s (severity: %s)\n", subArgs[1], proto, subArgs[2])

	case "remove":
		if len(subArgs) < 2 {
			return fmt.Errorf("usage: alert-rule remove <port/proto>")
		}
		port, proto, err := parseAlertPortProto(subArgs[1])
		if err != nil {
			return err
		}
		if !store.Remove(port, proto) {
			return fmt.Errorf("no alert rule found for %s", subArgs[1])
		}
		if err := store.Save(); err != nil {
			return fmt.Errorf("failed to save: %w", err)
		}
		fmt.Fprintf(os.Stdout, "alert rule removed for %s\n", subArgs[1])

	case "list":
		rules := store.All()
		if len(rules) == 0 {
			fmt.Fprintln(os.Stdout, "no alert rules defined")
			return nil
		}
		for _, r := range rules {
			fmt.Fprintf(os.Stdout, "%d/%s\t[%s]\t%s\n", r.Port, r.Protocol, r.Severity, r.Message)
		}

	default:
		return fmt.Errorf("unknown subcommand: %s", subArgs[0])
	}
	return nil
}

func parseAlertPortProto(s string) (int, string, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return 0, "", fmt.Errorf("invalid format %q: expected port/proto", s)
	}
	port, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("invalid port %q: %w", parts[0], err)
	}
	return port, parts[1], nil
}
