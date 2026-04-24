package alerting

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/portscanner"
)

// AlertLevel represents the severity of an alert.
type AlertLevel string

const (
	AlertInfo    AlertLevel = "INFO"
	AlertWarning AlertLevel = "WARNING"
	AlertCritical AlertLevel = "CRITICAL"
)

// Alert represents a single alerting event for an unexpected listener.
type Alert struct {
	Timestamp time.Time
	Level     AlertLevel
	Listener  portscanner.Listener
	Message   string
}

// Alerter handles formatting and dispatching alerts.
type Alerter struct {
	output    io.Writer
	allowlist map[uint16]struct{}
}

// NewAlerter creates an Alerter writing to the given output.
// allowedPorts is the set of ports considered expected/safe.
func NewAlerter(output io.Writer, allowedPorts []uint16) *Alerter {
	if output == nil {
		output = os.Stdout
	}
	allowlist := make(map[uint16]struct{}, len(allowedPorts))
	for _, p := range allowedPorts {
		allowlist[p] = struct{}{}
	}
	return &Alerter{output: output, allowlist: allowlist}
}

// Check compares current listeners against the allowlist and emits alerts
// for any unexpected listeners. Returns the list of alerts generated.
func (a *Alerter) Check(listeners []portscanner.Listener) []Alert {
	var alerts []Alert
	for _, l := range listeners {
		if _, ok := a.allowlist[l.Port]; !ok {
			alert := Alert{
				Timestamp: time.Now(),
				Level:     AlertWarning,
				Listener:  l,
				Message:   fmt.Sprintf("unexpected listener on %s:%d (pid %d)", l.Address, l.Port, l.PID),
			}
			alerts = append(alerts, alert)
			a.emit(alert)
		}
	}
	return alerts
}

// emit writes a formatted alert line to the configured output.
func (a *Alerter) emit(alert Alert) {
	fmt.Fprintf(
		a.output,
		"[%s] [%s] %s\n",
		alert.Timestamp.Format(time.RFC3339),
		alert.Level,
		alert.Message,
	)
}
