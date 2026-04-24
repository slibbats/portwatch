package daemon

import (
	"context"
	"log"
	"time"

	"github.com/user/portwatch/internal/alerting"
	"github.com/user/portwatch/internal/portscanner"
)

// Daemon periodically scans for port listeners and alerts on unexpected ones.
type Daemon struct {
	alerter  *alerting.Alerter
	interval time.Duration
	logger   *log.Logger
}

// New creates a new Daemon with the given alerter, scan interval, and logger.
func New(a *alerting.Alerter, interval time.Duration, logger *log.Logger) *Daemon {
	if logger == nil {
		logger = log.Default()
	}
	return &Daemon{
		alerter:  a,
		interval: interval,
		logger:   logger,
	}
}

// Run starts the daemon loop, scanning ports at each tick until ctx is cancelled.
func (d *Daemon) Run(ctx context.Context) error {
	d.logger.Printf("portwatch daemon started (interval: %s)", d.interval)
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	// Run an immediate scan before waiting for the first tick.
	if err := d.scan(); err != nil {
		d.logger.Printf("scan error: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := d.scan(); err != nil {
				d.logger.Printf("scan error: %v", err)
			}
		case <-ctx.Done():
			d.logger.Println("portwatch daemon stopped")
			return ctx.Err()
		}
	}
}

// scan performs a single port scan and passes results to the alerter.
func (d *Daemon) scan() error {
	listeners, err := portscanner.ScanListeners()
	if err != nil {
		return err
	}
	d.alerter.Check(listeners)
	return nil
}
