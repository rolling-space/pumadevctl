package internal

import (
	"errors"
	"fmt"
	"net"
	"sort"
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

// FindNextAvailablePortBlock returns the base port of the first available block within [min,max]
// where each allocation reserves `block` consecutive ports [base, base+block-1].
// Existing file entries are interpreted as reserving a block starting at their mapped port.
func FindNextAvailablePortBlock(entries []Entry, min, max, block int) (int, error) {
	if block <= 0 {
		return 0, fmt.Errorf("invalid block size: %d", block)
	}
	if min < 1 || max > 65535 || min > max {
		return 0, fmt.Errorf("invalid port range: %d-%d", min, max)
	}
	if max-min+1 < block {
		return 0, fmt.Errorf("port range too small for block size: range=%d, block=%d", max-min+1, block)
	}
	// collect reserved ranges
	type rng struct{ s, e int }
	var reserved []rng
	for _, e := range entries {
		if e.IsSymlink {
			continue
		}
		m, err := ParseMapping(e.Mapping)
		if err != nil {
			continue
		}
		base := m.Port
		reserved = append(reserved, rng{base, base + block - 1})
	}
	// merge (optional) - for faster checks sort by start
	sort.Slice(reserved, func(i, j int) bool { return reserved[i].s < reserved[j].s })
	// scan aligned to block boundaries for determinism
	for base := min; base+block-1 <= max; base += block {
		end := base + block - 1
		conflict := false
		for _, r := range reserved {
			if !(end < r.s || base > r.e) { // overlap
				conflict = true
				break
			}
		}
		if !conflict {
			return base, nil
		}
	}
	return 0, fmt.Errorf("no available port block in %d-%d with block size %d", min, max, block)
}
