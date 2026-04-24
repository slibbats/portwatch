package portscanner

import (
	"fmt"
	"io"
	"os"
	"sort"
)

// Summary holds aggregate statistics about a set of listeners.
type Summary struct {
	Total       int
	TCPCount    int
	UDPCount    int
	Loopback    int
	Public      int
	TopPorts    []uint16
}

// Summarize computes aggregate statistics from a list of listeners.
func Summarize(listeners []Listener) Summary {
	s := Summary{Total: len(listeners)}

	portFreq := make(map[uint16]int)
	for _, l := range listeners {
		switch l.Protocol {
		case "tcp", "tcp6":
			s.TCPCount++
		case "udp", "udp6":
			s.UDPCount++
		}
		if isLoopback(l.Address) {
			s.Loopback++
		} else {
			s.Public++
		}
		portFreq[l.Port]++
	}

	type portCount struct {
		port  uint16
		count int
	}
	var ranked []portCount
	for p, c := range portFreq {
		ranked = append(ranked, portCount{p, c})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].count != ranked[j].count {
			return ranked[i].count > ranked[j].count
		}
		return ranked[i].port < ranked[j].port
	})

	max := 5
	if len(ranked) < max {
		max = len(ranked)
	}
	for i := 0; i < max; i++ {
		s.TopPorts = append(s.TopPorts, ranked[i].port)
	}
	return s
}

// PrintSummary writes a human-readable summary to w (defaults to stdout).
func PrintSummary(s Summary, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	fmt.Fprintf(w, "Listeners : %d (TCP: %d, UDP: %d)\n", s.Total, s.TCPCount, s.UDPCount)
	fmt.Fprintf(w, "Loopback  : %d  Public: %d\n", s.Loopback, s.Public)
	if len(s.TopPorts) > 0 {
		fmt.Fprintf(w, "Top Ports : %v\n", s.TopPorts)
	}
}
