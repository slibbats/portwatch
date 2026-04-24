package alerting

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// Severity represents the alert severity level.
type Severity string

const (
	SeverityInfo    Severity = "INFO"
	SeverityWarning Severity = "WARNING"
	SeverityCritical Severity = "CRITICAL"
)

// Alert represents a single alert event.
type Alert struct {
	Timestamp time.Time
	Severity  Severity
	Message   string
	Listener  portscanner.Listener
}

// Notifier formats and writes alerts to an output destination.
type Notifier struct {
	out io.Writer
}

// NewNotifier creates a new Notifier. If out is nil, os.Stdout is used.
func NewNotifier(out io.Writer) *Notifier {
	if out == nil {
		out = os.Stdout
	}
	return &Notifier{out: out}
}

// Notify writes a formatted alert to the notifier's output.
func (n *Notifier) Notify(a Alert) error {
	_, err := fmt.Fprintf(
		n.out,
		"[%s] %s - %s (pid=%d addr=%s port=%d)\n",
		a.Timestamp.Format(time.RFC3339),
		a.Severity,
		a.Message,
		a.Listener.PID,
		a.Listener.Address,
		a.Listener.Port,
	)
	return err
}

// NotifyNew emits a WARNING alert for each newly detected listener.
func (n *Notifier) NotifyNew(listeners []portscanner.Listener) error {
	for _, l := range listeners {
		a := Alert{
			Timestamp: time.Now(),
			Severity:  SeverityWarning,
			Message:   "unexpected new listener detected",
			Listener:  l,
		}
		if err := n.Notify(a); err != nil {
			return err
		}
	}
	return nil
}

// NotifyGone emits an INFO alert for each listener that has disappeared.
func (n *Notifier) NotifyGone(listeners []portscanner.Listener) error {
	for _, l := range listeners {
		a := Alert{
			Timestamp: time.Now(),
			Severity:  SeverityInfo,
			Message:   "listener no longer active",
			Listener:  l,
		}
		if err := n.Notify(a); err != nil {
			return err
		}
	}
	return nil
}
