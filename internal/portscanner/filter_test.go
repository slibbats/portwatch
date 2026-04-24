package portscanner

import (
	"testing"
)

func makeFilterListener(ip, proto string, port int) Listener {
	return Listener{IP: ip, Port: port, Protocol: proto}
}

func TestFilterOptions_Apply_IncludeAll(t *testing.T) {
	f := DefaultFilterOptions()
	listeners := []Listener{
		makeFilterListener("0.0.0.0", "tcp", 80),
		makeFilterListener("127.0.0.1", "tcp", 8080),
		makeFilterListener("::1", "tcp6", 443),
	}
	got := f.Apply(listeners)
	if len(got) != 3 {
		t.Fatalf("expected 3 listeners, got %d", len(got))
	}
}

func TestFilterOptions_Apply_ExcludeLoopback(t *testing.T) {
	f := DefaultFilterOptions()
	f.IncludeLoopback = false
	listeners := []Listener{
		makeFilterListener("0.0.0.0", "tcp", 80),
		makeFilterListener("127.0.0.1", "tcp", 8080),
		makeFilterListener("::1", "tcp6", 443),
	}
	got := f.Apply(listeners)
	if len(got) != 1 {
		t.Fatalf("expected 1 listener after excluding loopback, got %d", len(got))
	}
	if got[0].Port != 80 {
		t.Errorf("expected port 80, got %d", got[0].Port)
	}
}

func TestFilterOptions_Apply_ExcludeIPv6(t *testing.T) {
	f := DefaultFilterOptions()
	f.IncludeIPv6 = false
	listeners := []Listener{
		makeFilterListener("0.0.0.0", "tcp", 80),
		makeFilterListener("::1", "tcp6", 443),
		makeFilterListener("::ffff:0.0.0.0", "tcp6", 8443),
	}
	got := f.Apply(listeners)
	if len(got) != 1 {
		t.Fatalf("expected 1 listener after excluding IPv6, got %d", len(got))
	}
	if got[0].Port != 80 {
		t.Errorf("expected port 80, got %d", got[0].Port)
	}
}

func TestFilterOptions_Apply_ProtocolFilter(t *testing.T) {
	f := DefaultFilterOptions()
	f.Protocols = []string{"udp"}
	listeners := []Listener{
		makeFilterListener("0.0.0.0", "tcp", 80),
		makeFilterListener("0.0.0.0", "udp", 53),
		makeFilterListener("0.0.0.0", "udp", 123),
	}
	got := f.Apply(listeners)
	if len(got) != 2 {
		t.Fatalf("expected 2 UDP listeners, got %d", len(got))
	}
}

func TestFilterOptions_Apply_EmptyInput(t *testing.T) {
	f := DefaultFilterOptions()
	got := f.Apply([]Listener{})
	if len(got) != 0 {
		t.Errorf("expected empty result, got %d", len(got))
	}
}
