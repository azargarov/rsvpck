package version

import (
    "runtime"
    "fmt"
)

var (
    Version    = "dev"
    GoVersion  = runtime.Version()
    BuildDate  = "unknown"
    CommitHash = "unknown"
)

func String() string {
    return fmt.Sprintf("rsvpck version %s (Go: %s, Build: %s, Commit: %s)", 
        Version, GoVersion, BuildDate, CommitHash)
}

func Short() string {
    return Version
}

func Info() map[string]string {
    return map[string]string{
        "version":    Version,
        "go_version": GoVersion,
        "build_date": BuildDate,
        "commit":     CommitHash,
    }
}