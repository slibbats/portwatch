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
// An initial snapshot is taken immediately on startup before the first tick,
// so that a baseline exists for diffing as soon as the first interval elapses.
func (sm *SnapshotManager) Run(ctx context.Context) {
	sm.opts.Logger.Printf("snapshot manager started (interval=%s)", sm.opts.Interval)
	ticker := time.NewTicker(sm.opts.Interval)
	defer ticker.Stop()
	defer close(sm.Diffs)

	// Capture an initial baseline snapshot so the first diff is meaningful.
	sm.takeBaseline()

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

// takeBaseline captures and persists the current listener state only when no
// prior snapshot exists, avoiding spurious diffs on restart.
func (sm *SnapshotManager) takeBaseline() {
	if _, err := sm.history.Latest(); err == nil {
		// A previous snapshot already exists; skip baseline capture.
		return
	}
	listeners, err := ScanListeners()
	if err != nil {
		sm.opts.Logger.Printf("baseline scan error: %v", err)
		return
	}
	if err := sm.history.Save(NewSnapshot(listeners)); err != nil {
		sm.opts.Logger.Printf("baseline save error: %v", err)
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
