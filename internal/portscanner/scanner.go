package portscanner

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Listener represents an active port listener on the system.
type Listener struct {
	Protocol string
	LocalAddress string
	LocalPort int
	PID int
	ProgramName string
}

// ScanListeners reads active TCP/UDP listeners from /proc/net/tcp and /proc/net/udp.
func ScanListeners() ([]Listener, error) {
	var listeners []Listener

	for _, proto := range []string{"tcp", "udp"} {
		path := fmt.Sprintf("/proc/net/%s", proto)
		entries, err := parseProcNet(path, proto)
		if err != nil {
			return nil, fmt.Errorf("scanning %s: %w", proto, err)
		}
		listeners = append(listeners, entries...)
	}

	return listeners, nil
}

// parseProcNet parses a /proc/net/tcp or /proc/net/udp file.
func parseProcNet(path, proto string) ([]Listener, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var listeners []Listener
	scanner := bufio.NewScanner(f)

	// Skip header line
	scanner.Scan()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// State 0A = LISTEN for TCP; for UDP state is always 07
		state := fields[3]
		if proto == "tcp" && state != "0A" {
			continue
		}

		localAddr := fields[1]
		addr, port, err := parseHexAddr(localAddr)
		if err != nil {
			continue
		}

		listeners = append(listeners, Listener{
			Protocol:     strings.ToUpper(proto),
			LocalAddress: addr,
			LocalPort:    port,
		})
	}

	return listeners, scanner.Err()
}

// parseHexAddr converts a hex-encoded address:port (e.g. "0100007F:0050") to dotted IP and int port.
func parseHexAddr(hexAddr string) (string, int, error) {
	parts := strings.SplitN(hexAddr, ":", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid address format: %s", hexAddr)
	}

	ipHex := parts[0]
	portHex := parts[1]

	ipInt, err := strconv.ParseUint(ipHex, 16, 32)
	if err != nil {
		return "", 0, err
	}

	ip := fmt.Sprintf("%d.%d.%d.%d",
		ipInt&0xFF,
		(ipInt>>8)&0xFF,
		(ipInt>>16)&0xFF,
		(ipInt>>24)&0xFF,
	)

	port, err := strconv.ParseInt(portHex, 16, 32)
	if err != nil {
		return "", 0, err
	}

	return ip, int(port), nil
}
