package portscanner

import (
	"context"
	"log"
	"time"
)

// SnapshotManagerOptions configures the snapshot manager behaviour.
type SnapshotManagerOptions struct {
	Dir      string
	Interval time.Duration
	MaxFiles int
	Logger   *log.Logger
}

// DefaultSnapshotManagerOptions returns sensible defaults.
func DefaultSnapshotManagerOptions(dir string) SnapshotManagerOptions {
	return SnapshotManagerOptions{
		Dir:      dir,
		Interval: 60 * time.Second,
		MaxFiles: 48,
		Logger:   log.Default(),
	}
}

// SnapshotManager periodically captures port snapshots, persists them, and
// emits a DiffResult whenever the listener set changes.
type SnapshotManager struct {
	opts    SnapshotManagerOptions
	history *HistoryStore
	Diffs   chan DiffResult
}

// NewSnapshotManager creates a SnapshotManager using the provided options.
func NewSnapshotManager(opts SnapshotManagerOptions) (*SnapshotManager, error) {
	hs, err := NewHistoryStore(opts.Dir, opts.MaxFiles)
	if err != nil {
		return nil, err
	}
	return &SnapshotManager{
		opts:    opts,
		history: hs,
		Diffs:   make(chan DiffResult, 8),
	}, nil
}

// Run starts the periodic snapshot loop. It blocks until ctx is cancelled.
func (sm *SnapshotManager) Run(ctx context.Context) {
	sm.opts.Logger.Printf("snapshot manager started (interval=%s)", sm.opts.Interval)
	ticker := time.NewTicker(sm.opts.Interval)
	defer ticker.Stop()
	defer close(sm.Diffs)

	for {
		select {
		case <-ctx.Done():
			sm.opts.Logger.Println("snapshot manager stopped")
			return
		case <-ticker.C:
			sm.tick()
		}
	}
}

func (sm *SnapshotManager) tick() {
	listeners, err := ScanListeners()
	if err != nil {
		sm.opts.Logger.Printf("snapshot scan error: %v", err)
		return
	}

	curr := NewSnapshot(listeners)

	if prev, err := sm.history.Latest(); err == nil {
		if diff := DiffSnapshots(prev, curr); !diff.IsEmpty() {
			sm.opts.Logger.Printf("snapshot diff: %s", diff.Summary())
			select {
			case sm.Diffs <- diff:
			default:
			}
		}
	}

	if err := sm.history.Save(curr); err != nil {
		sm.opts.Logger.Printf("snapshot save error: %v", err)
	}
}
