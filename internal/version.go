package internal

import (
	"fmt"
	"runtime"
)

// Version metadata populated at build time via -ldflags. Defaults are for dev builds.
var (
	// Version is the semver for the release. e.g. v0.4.0. Defaults to "dev".
	Version = "dev"
	// Commit is the short git commit SHA for reproducibility.
	Commit = ""
	// Date is the build timestamp in UTC (ISO-8601), e.g. 2025-10-20T04:21:00Z.
	Date = ""
)

// VersionSummary returns a human-friendly single-line version suitable for Cobra's Version field.
func VersionSummary() string {
	base := Version
	meta := ""
	if Commit != "" {
		meta = Commit
	}
	if Date != "" {
		if meta != "" {
			meta = fmt.Sprintf("%s, %s", meta, Date)
		} else {
			meta = Date
		}
	}
	if meta != "" {
		return fmt.Sprintf("%s (%s, %s/%s)", base, meta, runtime.GOOS, runtime.GOARCH)
	}
	return fmt.Sprintf("%s (%s/%s)", base, runtime.GOOS, runtime.GOARCH)
}

// VersionLong returns a multi-line detailed version string.
func VersionLong() string {
	return fmt.Sprintf("version: %s\ncommit: %s\ndate:   %s\ngo:     %s %s/%s",
		Version,
		valueOrDash(Commit),
		valueOrDash(Date),
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}

func valueOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
