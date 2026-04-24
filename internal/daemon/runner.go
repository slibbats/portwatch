package daemon

import (
	"context"
	"log"

	"github.com/user/portwatch/internal/alerting"
	"github.com/user/portwatch/internal/portscanner"
)

// Runner ties together a Watcher and a Notifier to form the main watch loop.
type Runner struct {
	watcher  *portscanner.Watcher
	notifier *alerting.Notifier
	logger   *log.Logger
}

// NewRunner constructs a Runner from a baseline, watch options, and notifier.
func NewRunner(
	baseline []portscanner.Listener,
	opts portscanner.WatchOptions,
	notifier *alerting.Notifier,
	logger *log.Logger,
) *Runner {
	return &Runner{
		watcher:  portscanner.NewWatcher(baseline, opts),
		notifier: notifier,
		logger:   logger,
	}
}

// Run starts the watch loop and blocks until ctx is cancelled.
func (r *Runner) Run(ctx context.Context) {
	r.logger.Println("[runner] starting watch loop")
	results := r.watcher.Watch(ctx)
	for result := range results {
		if len(result.New) > 0 {
			r.logger.Printf("[runner] %d new listener(s) detected", len(result.New))
			r.notifier.NotifyNew(result.New)
		}
		if len(result.Gone) > 0 {
			r.logger.Printf("[runner] %d listener(s) disappeared", len(result.Gone))
		}
	}
	r.logger.Println("[runner] watch loop stopped")
}
