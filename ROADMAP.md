# Roadmap

## v0.4.0
- Configurable endpoints/targets via CLI flag (`--config path`) in addition to embedded defaults.
- Pluggable output “profiles” (compact vs verbose detail/errors).
- Optional export of raw probe JSON for dashboards/logging. 
- Extended TLS diagnostics (chain summary per hop). 
  
## v0.3.x
- Remove legacy `--policy` flag and dead code paths. 
- Tune default worker-pool settings and expose via env/flags. 
- Add basic JSON renderer for machine parsing (optional third output mode). {index=15}
- Expand tests for adapters (HTTP proxy auth paths, TLS errors, ICMP command detection). 

## v0.2.0 — DDD foundation + CLI UX
- Domain model: 
- Hexagonal architecture: ports/adapters for DNS, HTTP, ICMP, Proxy
- Output model: renderers (table/text/json) as adapters.
- Concurrency model: worker pool for checks; timeouts & cancellation (context).
- Collect CRM info
- Packaging: smaller binaries via `-ldflags`
- Compress: upx --best --lzma rsvpck.exe
 