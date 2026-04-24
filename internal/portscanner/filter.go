package portscanner

import "net"

// FilterOptions controls which listeners are included in scan results.
type FilterOptions struct {
	// IncludeLoopback includes listeners bound to 127.0.0.1 or ::1.
	IncludeLoopback bool
	// IncludeIPv6 includes listeners on IPv6 addresses.
	IncludeIPv6 bool
	// Protocols restricts results to the given protocols (e.g. "tcp", "udp").
	// An empty slice means all protocols are included.
	Protocols []string
}

// DefaultFilterOptions returns permissive filter options suitable for most
// monitoring scenarios.
func DefaultFilterOptions() FilterOptions {
	return FilterOptions{
		IncludeLoopback: true,
		IncludeIPv6:     true,
		Protocols:       []string{},
	}
}

// Apply returns only the listeners from the provided slice that satisfy the
// filter options.
func (f FilterOptions) Apply(listeners []Listener) []Listener {
	var result []Listener
	for _, l := range listeners {
		if !f.matchesProtocol(l.Protocol) {
			continue
		}
		if !f.IncludeLoopback && isLoopback(l.IP) {
			continue
		}
		if !f.IncludeIPv6 && isIPv6(l.IP) {
			continue
		}
		result = append(result, l)
	}
	return result
}

func (f FilterOptions) matchesProtocol(proto string) bool {
	if len(f.Protocols) == 0 {
		return true
	}
	for _, p := range f.Protocols {
		if p == proto {
			return true
		}
	}
	return false
}

func isLoopback(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return parsed.IsLoopback()
}

func isIPv6(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return parsed.To4() == nil
}
