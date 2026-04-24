package daemon

import (
	"bytes"
	"context"
	"log"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alerting"
)

func newTestDaemon(buf *bytes.Buffer) *Daemon {
	cfg := alerting.DefaultConfig()
	logger := log.New(buf, "", 0)
	alerter := alerting.NewAlerter(cfg, buf)
	return New(alerter, 50*time.Millisecond, logger)
}

func TestDaemon_New_DefaultLogger(t *testing.T) {
	cfg := alerting.DefaultConfig()
	alerter := alerting.NewAlerter(cfg, nil)
	d := New(alerter, time.Second, nil)
	if d.logger == nil {
		t.Fatal("expected non-nil logger when nil is passed")
	}
	if d.interval != time.Second {
		t.Errorf("expected interval 1s, got %s", d.interval)
	}
}

func TestDaemon_Run_StopsOnContextCancel(t *testing.T) {
	var buf bytes.Buffer
	d := newTestDaemon(&buf)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- d.Run(ctx)
	}()

	// Allow at least one tick to fire.
	time.Sleep(120 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("daemon did not stop after context cancellation")
	}
}

func TestDaemon_Run_LogsStartAndStop(t *testing.T) {
	var buf bytes.Buffer
	d := newTestDaemon(&buf)

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()

	_ = d.Run(ctx)

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected log output, got none")
	}
}
