package portscanner

import (
	"context"
	"log"
	"time"
)

// WatchOptions configures the port watcher behavior.
type WatchOptions struct {
	Interval      time.Duration
	FilterOptions FilterOptions
	Logger        *log.Logger
}

// DefaultWatchOptions returns sensible defaults for watching.
func DefaultWatchOptions(logger *log.Logger) WatchOptions {
	return WatchOptions{
		Interval:      15 * time.Second,
		FilterOptions: DefaultFilterOptions(),
		Logger:        logger,
	}
}

// WatchResult holds the outcome of a single scan cycle.
type WatchResult struct {
	New     []Listener
	Gone    []Listener
	Scanned []Listener
}

// Watcher periodically scans for port changes against a baseline.
type Watcher struct {
	baseline map[string]Listener
	opts     WatchOptions
}

// NewWatcher creates a Watcher with the given baseline and options.
func NewWatcher(baseline []Listener, opts WatchOptions) *Watcher {
	bm := make(map[string]Listener, len(baseline))
	for _, l := range baseline {
		bm[listenerKey(l)] = l
	}
	return &Watcher{baseline: bm, opts: opts}
}

// Watch runs the watch loop, sending results on the returned channel.
// It stops when ctx is cancelled.
func (w *Watcher) Watch(ctx context.Context) <-chan WatchResult {
	results := make(chan WatchResult, 1)
	go func() {
		defer close(results)
		ticker := time.NewTicker(w.opts.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				result, err := w.scan()
				if err != nil {
					w.opts.Logger.Printf("[watcher] scan error: %v", err)
					continue
				}
				select {
				case results <- result:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return results
}

func (w *Watcher) scan() (WatchResult, error) {
	listeners, err := ScanListeners()
	if err != nil {
		return WatchResult{}, err
	}
	filtered := w.opts.FilterOptions.Apply(listeners)
	current := make(map[string]Listener, len(filtered))
	for _, l := range filtered {
		current[listenerKey(l)] = l
	}
	var newL, goneL []Listener
	for k, l := range current {
		if _, ok := w.baseline[k]; !ok {
			newL = append(newL, l)
		}
	}
	for k, l := range w.baseline {
		if _, ok := current[k]; !ok {
			goneL = append(goneL, l)
		}
	}
	return WatchResult{New: newL, Gone: goneL, Scanned: filtered}, nil
}
