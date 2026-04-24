package daemon

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alerting"
	"github.com/user/portwatch/internal/portscanner"
)

func newTestRunner(buf *bytes.Buffer) *Runner {
	logger := log.New(buf, "", 0)
	notifier := alerting.NewNotifier(buf)
	opts := portscanner.DefaultWatchOptions(logger)
	opts.Interval = 20 * time.Millisecond
	return NewRunner(nil, opts, notifier, logger)
}

func TestRunner_Run_StopsOnContextCancel(t *testing.T) {
	var buf bytes.Buffer
	r := newTestRunner(&buf)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	done := make(chan struct{})
	go func() {
		r.Run(ctx)
		close(done)
	}()
	select {
	case <-done:
		// expected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run did not stop after context cancellation")
	}
}

func TestRunner_Run_LogsStartAndStop(t *testing.T) {
	var buf bytes.Buffer
	r := newTestRunner(&buf)
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	r.Run(ctx)
	out := buf.String()
	if !strings.Contains(out, "starting watch loop") {
		t.Errorf("expected start log, got: %s", out)
	}
	if !strings.Contains(out, "watch loop stopped") {
		t.Errorf("expected stop log, got: %s", out)
	}
}

func TestNewRunner_NotNil(t *testing.T) {
	var buf bytes.Buffer
	r := newTestRunner(&buf)
	if r == nil {
		t.Fatal("expected non-nil runner")
	}
	if r.watcher == nil {
		t.Error("expected watcher to be initialised")
	}
	if r.notifier == nil {
		t.Error("expected notifier to be initialised")
	}
}
