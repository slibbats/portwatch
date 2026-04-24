package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/portwatch/internal/alerting"
	"github.com/user/portwatch/internal/daemon"
)

func main() {
	interval := flag.Duration("interval", 5*time.Second, "how often to scan for port changes")
	configPath := flag.String("config", "", "path to config file (optional)")
	flag.Parse()

	logger := log.New(os.Stdout, "[portwatch] ", log.LstdFlags)

	var cfg alerting.Config
	var err error

	if *configPath != "" {
		cfg, err = alerting.LoadConfig(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
			os.Exit(1)
		}
	} else {
		cfg = alerting.DefaultConfig()
	}

	alerter := alerting.NewAlerter(cfg, os.Stdout)
	d := daemon.New(alerter, *interval, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := d.Run(ctx); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "daemon exited with error: %v\n", err)
		os.Exit(1)
	}
}
