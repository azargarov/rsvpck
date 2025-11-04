# rsvpck — RSvP Connectivity Checker

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**rsvpck** is a tiny, single-binary network diagnostic that checks connectivity across **Direct Internet**, **Proxy**, and **VPN** paths using concurrent probes (ICMP, DNS, TCP, HTTP, TLS). Results are rendered as a compact table (default) or plain text. 

## Features

- **Parallel probes** with retry/backoff for transient failures (I/O-bound, fast end-to-end). 
- **Direct / Proxy / VPN** endpoint groups + automatic **mode** detection (Direct / Via Proxy / Via VPN / None). :contentReference[oaicite:25]{index=25}
- **TLS certificate fetch** with smart fallback via VPN proxy list. 
- **Friendly output**: table (default) or text; smart ASCII/Unicode symbols. 
- **Embedded defaults** (YAML/JSON) so it “just works” out of the box. 
- **CI releases**: Linux & Windows artifacts, UPX-compressed, with SHA256SUMS. {index=29}

## Quick start

Download a release artifact for your OS from GitHub Releases and run:

```bash
./rsvpck           # table renderer (default)
./rsvpck --text    # plain text renderer
./rsvpck --ascii   # force ASCII, no Unicode
./rsvpck --version # print version/build info
