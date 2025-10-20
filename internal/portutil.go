package internal

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type Mapping struct {
	Host string
	Port int
	Raw  string // original string
}

// ParseMapping accepts "3000", "127.0.0.1:3000", "localhost:3000", "[::1]:3000"
func ParseMapping(s string) (*Mapping, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty mapping")
	}
	// plain port
	if _, err := strconv.Atoi(s); err == nil {
		p, _ := strconv.Atoi(s)
		if p < 1 || p > 65535 {
			return nil, fmt.Errorf("invalid port %d", p)
		}
		return &Mapping{Host: "127.0.0.1", Port: p, Raw: s}, nil
	}
	host, portStr, err := net.SplitHostPort(s)
	if err != nil {
		return nil, fmt.Errorf("invalid host:port mapping: %w", err)
	}
	p, err := strconv.Atoi(portStr)
	if err != nil || p < 1 || p > 65535 {
		return nil, fmt.Errorf("invalid port %q", portStr)
	}
	if host == "" {
		host = "127.0.0.1"
	}
	return &Mapping{Host: host, Port: p, Raw: s}, nil
}

// IsPortReachable tries to connect to host:port with timeout
func IsPortReachable(host string, port int, timeout time.Duration) bool {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// FindNextFreePort scans from start, skipping used list and testing bindability
func FindNextFreePort(start int, used map[int]bool) (int, error) {
	if start < 1024 {
		start = 1024
	}
	for p := start; p <= 65535; p++ {
		if used[p] {
			continue
		}
		ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(p)))
		if err != nil {
			continue
		}
		_ = ln.Close()
		return p, nil
	}
	return 0, fmt.Errorf("no free port found")
}
