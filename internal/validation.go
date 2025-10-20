package internal

import (
    "time"
)

type ValidationResult struct {
    Entry
    Reachable bool   `json:"reachable"`
    Reason    string `json:"reason,omitempty"`
}

// ValidateEntries checks TCP reachability for non-symlink entries
func ValidateEntries(entries []Entry, timeoutMs int) []ValidationResult {
    timeout := time.Duration(timeoutMs) * time.Millisecond
    results := make([]ValidationResult, 0, len(entries))
    for _, e := range entries {
        vr := ValidationResult{Entry: e, Reachable: true}
        if e.IsSymlink {
            // optional: could check if target exists
            results = append(results, vr)
            continue
        }
        m, err := ParseMapping(e.Mapping)
        if err != nil {
            vr.Reachable = false
            vr.Reason = err.Error()
            results = append(results, vr)
            continue
        }
        if !IsPortReachable(m.Host, m.Port, timeout) {
            vr.Reachable = false
            vr.Reason = "connection failed"
        }
        results = append(results, vr)
    }
    return results
}
