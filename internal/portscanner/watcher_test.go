package portscanner

import (
	"context"
	"log"
	"os"
	"testing"
	"time"
)

func testLogger() *log.Logger {
	return log.New(os.Stderr, "", 0)
}

func makeWatchListener(proto, addr string, port uint16) Listener {
	return Listener{Protocol: proto, Address: addr, Port: port, PID: 0, Process: "test"}
}

func TestNewWatcher_StoresBaseline(t *testing.T) {
	baseline := []Listener{
		makeWatchListener("tcp", "0.0.0.0", 80),
		makeWatchListener("tcp", "0.0.0.0", 443),
	}
	opts := DefaultWatchOptions(testLogger())
	w := NewWatcher(baseline, opts)
	if len(w.baseline) != 2 {
		t.Fatalf("expected 2 baseline entries, got %d", len(w.baseline))
	}
}

func TestWatcher_Scan_DetectsNew(t *testing.T) {
	// Empty baseline so any real listener appears as "new".
	opts := DefaultWatchOptions(testLogger())
	w := NewWatcher(nil, opts)
	result, err := w.scan()
	if err != nil {
		// On systems without /proc/net this may fail; skip gracefully.
		t.Skipf("scan not supported on this platform: %v", err)
	}
	// All scanned listeners should appear as new since baseline is empty.
	if len(result.New) != len(result.Scanned) {
		t.Errorf("expected new=%d to equal scanned=%d", len(result.New), len(result.Scanned))
	}
	if len(result.Gone) != 0 {
		t.Errorf("expected no gone listeners, got %d", len(result.Gone))
	}
}

func TestWatcher_Watch_ClosesOnCancel(t *testing.T) {
	opts := DefaultWatchOptions(testLogger())
	opts.Interval = 50 * time.Millisecond
	w := NewWatcher(nil, opts)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	ch := w.Watch(ctx)
	// Drain results until channel closes.
	for range ch {
	}
	select {
	case _, open := <-ch:
		if open {
			t.Error("expected channel to be closed after context cancel")
		}
	default:
	}
}

func TestDefaultWatchOptions(t *testing.T) {
	l := testLogger()
	opts := DefaultWatchOptions(l)
	if opts.Interval != 15*time.Second {
		t.Errorf("expected 15s interval, got %v", opts.Interval)
	}
	if opts.Logger != l {
		t.Error("expected logger to be set")
	}
}
