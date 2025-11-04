# rsvpck — RSvP Connectivity Checker

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

[![CI & Release](https://github.com/azargarov/rsvpck/actions/workflows/ci-release.yml/badge.svg)](https://github.com/azargarov/rsvpck/actions/workflows/ci.yaml)


**rsvpck** is a tiny, single-binary network diagnostic that checks connectivity across **Direct Internet**, **Proxy**, and **VPN** paths using concurrent probes (ICMP, DNS, TCP, HTTP, TLS). Results are rendered as a compact table (default) or plain text. 

## Features

- **Parallel probes** with retry/backoff for transient failures (I/O-bound, fast end-to-end). 
- **Direct / Proxy / VPN** endpoint groups + automatic **mode** detection (Direct / Via Proxy / Via VPN / None). 
- **TLS certificate fetch** with smart fallback via VPN proxy list. 
- **Friendly output**: table (default) or text; smart ASCII/Unicode symbols. 
- **Embedded defaults** (YAML/JSON) so it “just works” out of the box. 
- **CI releases**: Linux & Windows artifacts, UPX-compressed, with SHA256SUMS. 

## Quick start

Download a release artifact for your OS from GitHub Releases and run:

```bash
./rsvpck           # table renderer (default)
./rsvpck --text    # plain text renderer
./rsvpck --ascii   # force ASCII, no Unicode
./rsvpck --version # print version/build info
```

##  Core Design Principles

- **Isolation:** The `domain` layer has no external dependencies.  
- **Concurrency:** The `app.Executor` orchestrates concurrent probe jobs using a reusable worker-pool with retry/backoff.  
- **Extensibility:** Each probe type is a pluggable adapter implementing a common interface.  
- **Resilience:** Transient network errors trigger exponential backoff and retry policies.  
- **Observability:** Structured logging (`zlog`) and contextual traces make each probe’s result transparent.  
- **Portability:** Built as a single statically-linked binary for Linux and Windows, requiring no dependencies.

## Execution Flow

1. CLI parses user options and reads config.
2. `Executor` creates probe jobs for each endpoint group.  
3. Each probe runs concurrently via the worker-pool.  
4. Results are aggregated into `ConnectivityResult`.  
5. Renderer outputs table or text summary.  
6. Exit code reflects overall connectivity status.

