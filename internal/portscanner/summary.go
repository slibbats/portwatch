package portscanner

import (
	"fmt"
	"io"
	"os"
	"sort"
)

// Summary holds aggregated statistics about a set of listeners.
type Summary struct {
	TotalCount      int
	TCPCount        int
	UDPCount        int
	UniqueProcesses int
	TopProcesses    []ProcessCount
}

// ProcessCount pairs a process name with how many ports it holds open.
type ProcessCount struct {
	Process string
	Count   int
}

// Summarize computes a Summary from a slice of Listener values.
func Summarize(listeners []Listener) Summary {
	processCounts := make(map[string]int)
	var tcpCount, udpCount int

	for _, l := range listeners {
		switch l.Protocol {
		case "tcp", "tcp6":
			tcpCount++
		case "udp", "udp6":
			udpCount++
		}
		name := l.Process
		if name == "" {
			name = "unknown"
		}
		processCounts[name]++
	}

	top := make([]ProcessCount, 0, len(processCounts))
	for proc, cnt := range processCounts {
		top = append(top, ProcessCount{Process: proc, Count: cnt})
	}
	sort.Slice(top, func(i, j int) bool {
		if top[i].Count != top[j].Count {
			return top[i].Count > top[j].Count
		}
		return top[i].Process < top[j].Process
	})

	return Summary{
		TotalCount:      len(listeners),
		TCPCount:        tcpCount,
		UDPCount:        udpCount,
		UniqueProcesses: len(processCounts),
		TopProcesses:    top,
	}
}

// PrintSummary writes a human-readable summary of listeners to w.
// If w is nil, output goes to os.Stdout.
func PrintSummary(listeners []Listener, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}

	s := Summarize(listeners)

	fmt.Fprintf(w, "=== Listener Summary ===\n")
	fmt.Fprintf(w, "Total listeners : %d\n", s.TotalCount)
	fmt.Fprintf(w, "TCP             : %d\n", s.TCPCount)
	fmt.Fprintf(w, "UDP             : %d\n", s.UDPCount)
	fmt.Fprintf(w, "Unique Process  : %d\n", s.UniqueProcesses)

	if len(s.TopProcesses) > 0 {
		fmt.Fprintf(w, "\nProcess breakdown:\n")
		for _, pc := range s.TopProcesses {
			fmt.Fprintf(w, "  %-20s %d\n", pc.Process, pc.Count)
		}
	}
}
