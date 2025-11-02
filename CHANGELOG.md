# Changelog


All notable changes to this project will be documented in this file.

## [v0.3.1] — 2025-11-02

### Added
- Centralized versioning via `internal/version` package.
- `httpx` adapter migrated to use the version package.

### Changed
- Cursor visibility now toggles **only** when attached to a TTY (safer UX in pipes/CI).

### Tests
- Increased coverage across domain, app (executor), adapters (`httpx`, text renderer), and config loaders.

### Docs
- Fixed typos; updated CHANGELOG and ROADMAP.

**Notes:** Default testing policy remains **Exhaustive** for consistency.

## [v0.3.0] — 2025-10-26

### Added
- **Concurrent execution via worker pool** — all probes (ICMP, DNS, TCP, HTTP, TLS) now run in parallel using a unified pool.
- **Retry & backoff logic** for transient I/O failures, with exponential backoff and per-job retry policies.
- **Spinner** while collecting results for better UX.

### Changed
- **Removed execution policies.**  
  The previous `Optimized` vs `Exhaustive` distinction is deprecated.  
  All probes now execute concurrently since the workload is I/O-bound and benefits from higher parallelism.
- **Default pool size increased** to 15 workers (configurable).  
  Since probes spend most time waiting on network I/O, worker count can safely exceed CPU cores (hard coded max 25).
- **Simplified executor flow** — no ICMP gating; all endpoints are tested in a single pass for faster overall runtime.
- **Improved probe status handling** — `MarkFailure` now preserves detailed classifier statuses (Timeout, DNSFailure, etc.).
- **Go 1.24 syntax updates** — uses modern `for range n` loops and streamlined concurrency patterns.
- **Renderers**: refined table/text output, ASCII/Unicode handling.

### Fixed
- Corrected race-safe worker accounting using atomic counters.
- Properly handle context cancellation and job cleanup on shutdown.

### Notes
- The `--policy` flag is ignored and will be removed in a future release.
- v0.3.0 focuses on correctness, simplicity, and performance; smarter scheduling logic may return in v0.4.x.


## [v0.2.0] — 2025-10-19

### Added
- **Renderer / output model** with pluggable **text** and **table** renderers.
- **Execution policy: Optimized** — skip remaining probes when early checks fail (per mode).
- **Custom error code module** for DNS/HTTP/ICMP.
- **Config loading** from external files and embedded FS (defaults included).
- **Render configuration** knobs (formatting helpers).
- **Spinner/animation** while gathering information.
- **Version flag & render flags** (force ASCII, force renderer).
- **Speed test** scaffold for future use.
- **TLS certificate fetch via proxy/VPN** when direct internet is not available.

### Changed
- **Endpoints grouped** into Direct / VPN / Proxy sets; analyzer adapted.
- **Default policy** switched to **Exhaustive**.
- **Config moved to YAML** (plus GE defaults).
- **Unicode detection & render refactor**; optional forced ASCII headers/characters.
- **Timeouts increased** (TCP/TLS dialer ~10s).
- **Module metadata** updated (`go.mod`, `go.sum`).

### Fixed
- **TLS certificate fetching** reliability and context timeout handling (plus waiting spinner for slow TLS).
- Corrected **IP addresses** in defaults.
- Adjusted **result interpretation** logic in analyzer.
- Fixed **CRM-number collection** command in host info.

### CI
- **Race detector** (CGO enabled) and CI fixes; **v0.2**-aligned workflow; temporary disable/enable phases during migration; UPX compression retained from earlier work.

### Notes
- Initial v0.2 entry existed with high-level bullets; this update consolidates **all v0.2 work** up to Oct 19, 2025 into a single release, as requested.

[Unreleased]: https://github.com/azargarov/rsvpck/compare/v0.2.0...HEAD


## [v0.2.0] - 2025-10-07
- Domain model
- Hexagonal architecture: ports/adapters for DNS, HTTP, ICMP, Proxy
- Output model: renderers table/text as adapters.
- Config file for probes