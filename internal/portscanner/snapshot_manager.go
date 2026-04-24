package portscanner

import (
	"context"
	"log"
	"time"
)

// SnapshotManagerOptions configures periodic snapshot behaviour.
type SnapshotManagerOptions struct {
	Interval time.Duration
	MaxAge   time.Duration
	Logger   *log.Logger
}

// DefaultSnapshotManagerOptions returns sensible defaults.
func DefaultSnapshotManagerOptions() SnapshotManagerOptions {
	return SnapshotManagerOptions{
		Interval: 5 * time.Minute,
		MaxAge:   7 * 24 * time.Hour,
	}
}

// SnapshotManager periodically captures listener snapshots and stores them.
type SnapshotManager struct {
	store   *HistoryStore
	filter  FilterOptions
	opts    SnapshotManagerOptions
	logger  *log.Logger
}

// NewSnapshotManager creates a SnapshotManager writing to store.
func NewSnapshotManager(store *HistoryStore, filter FilterOptions, opts SnapshotManagerOptions) *SnapshotManager {
	logger := opts.Logger
	if logger == nil {
		logger = log.Default()
	}
	return &SnapshotManager{
		store:  store,
		filter: filter,
		opts:   opts,
		logger: logger,
	}
}

// Run captures snapshots on the configured interval until ctx is cancelled.
func (m *SnapshotManager) Run(ctx context.Context) error {
	ticker := time.NewTicker(m.opts.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			m.logger.Println("snapshot manager: stopping")
			return ctx.Err()
		case <-ticker.C:
			if err := m.capture(); err != nil {
				m.logger.Printf("snapshot manager: capture error: %v", err)
			}
			if m.opts.MaxAge > 0 {
				if err := m.store.Prune(m.opts.MaxAge); err != nil {
					m.logger.Printf("snapshot manager: prune error: %v", err)
				}
			}
		}
	}
}

func (m *SnapshotManager) capture() error {
	listeners, err := ScanListeners()
	if err != nil {
		return err
	}
	filtered := m.filter.Apply(listeners)
	snap := NewSnapshot(filtered)
	if err := m.store.Save(snap); err != nil {
		return err
	}
	m.logger.Printf("snapshot manager: saved snapshot with %d listeners", len(filtered))
	return nil
}
