package alerting_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/alerting"
	"github.com/user/portwatch/internal/portscanner"
)

func makeListener(addr string, port uint16, pid int) portscanner.Listener {
	return portscanner.Listener{Address: addr, Port: port, PID: pid}
}

func TestAlerter_NoAlertsForAllowedPorts(t *testing.T) {
	var buf bytes.Buffer
	a := alerting.NewAlerter(&buf, []uint16{80, 443})

	listeners := []portscanner.Listener{
		makeListener("0.0.0.0", 80, 100),
		makeListener("0.0.0.0", 443, 101),
	}

	alerts := a.Check(listeners)
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts, got %d", len(alerts))
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output, got: %s", buf.String())
	}
}

func TestAlerter_AlertForUnexpectedPort(t *testing.T) {
	var buf bytes.Buffer
	a := alerting.NewAlerter(&buf, []uint16{80})

	listeners := []portscanner.Listener{
		makeListener("0.0.0.0", 80, 100),
		makeListener("127.0.0.1", 9999, 202),
	}

	alerts := a.Check(listeners)
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Listener.Port != 9999 {
		t.Errorf("expected alert for port 9999, got %d", alerts[0].Listener.Port)
	}
	if alerts[0].Level != alerting.AlertWarning {
		t.Errorf("expected WARNING level, got %s", alerts[0].Level)
	}
	output := buf.String()
	if !strings.Contains(output, "9999") {
		t.Errorf("expected port 9999 in output, got: %s", output)
	}
}

func TestAlerter_MultipleUnexpected(t *testing.T) {
	var buf bytes.Buffer
	a := alerting.NewAlerter(&buf, []uint16{})

	listeners := []portscanner.Listener{
		makeListener("0.0.0.0", 22, 1),
		makeListener("0.0.0.0", 8080, 2),
	}

	alerts := a.Check(listeners)
	if len(alerts) != 2 {
		t.Errorf("expected 2 alerts, got %d", len(alerts))
	}
}

func TestNewAlerter_NilOutputDefaultsToStdout(t *testing.T) {
	// Should not panic when output is nil
	a := alerting.NewAlerter(nil, []uint16{80})
	if a == nil {
		t.Error("expected non-nil Alerter")
	}
}
