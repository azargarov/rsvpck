package config_test

import (
	"testing"
	"os"
	"path/filepath"

	"github.com/azargarov/rsvpck/internal/config"
	"github.com/azargarov/rsvpck/internal/domain"
)

const sampleYAML = `
proxyURL: http://proxy.local:8080
directEndpoints:
  - { target: "1.1.1.1", type: public, kind: icmp, note: "ping" }
  - { target: "example.com", type: public, kind: dns, note: "dns" }
  - { target: "example.com:443", type: public, kind: tcp, note: "tcp" }
  - { target: "https://example.com", type: public, kind: http, note: "http", useProxy: false }
proxyEndpoints:
  - { target: "https://example.com", type: public, kind: http, note: "http via proxy", useProxy: true }
vpnEndpoints:
  - { target: "10.0.0.5:443", type: vpn, kind: tcp, note: "vpn tcp" }
`

const sampleJSON = `{
  "proxyURL": "http://proxy.local:8080",
  "directEndpoints": [
    {"target":"1.1.1.1","type":"public","kind":"icmp","note":"ping"},
    {"target":"example.com","type":"public","kind":"dns","note":"dns"},
    {"target":"example.com:443","type":"public","kind":"tcp","note":"tcp"},
    {"target":"https://example.com","type":"public","kind":"http","note":"http","useProxy":false}
  ],
  "proxyEndpoints": [
    {"target":"https://example.com","type":"public","kind":"http","note":"http via proxy","useProxy":true}
  ],
  "vpnEndpoints": [
    {"target":"10.0.0.5:443","type":"vpn","kind":"tcp","note":"vpn tcp"}
  ]
}`

func TestParseConfigBytes_YAML(t *testing.T) {
	cfg, err := callParse(sampleYAML, ".yaml")
	if err != nil {
		t.Fatalf("parse yaml: %v", err)
	}
	assertConfigShape(t, cfg)
}

func TestParseConfigBytes_JSON(t *testing.T) {
	cfg, err := callParse(sampleJSON, ".json")
	if err != nil {
		t.Fatalf("parse json: %v", err)
	}
	assertConfigShape(t, cfg)
}

func callParse(s, ext string) (domain.NetTestConfig, error) {
	// access unexported parseConfigBytes via exported wrappers is not available;
	// so we simulate writing a temp file and using LoadFromFile.
	// Simpler: re-use specToDomain path by parsing with exts through file-like path.
	// We’ll write to a tmp file here (no network).
	return configFor(ext, s)
}

func configFor(ext, body string) (domain.NetTestConfig, error) {
	dir := testingTempDir()
	f := filepath.Join(dir, "cfg"+ext)
	_ = os.WriteFile(f, []byte(body), 0o644)
	return config.LoadFromFile(f)
}

func testingTempDir() string {
	dir, _ := os.MkdirTemp("", "rsvpck-")
	return dir
}

func assertConfigShape(t *testing.T, cfg domain.NetTestConfig) {
	t.Helper()
	if cfg.ProxyURL == "" {
		t.Fatalf("missing ProxyURL")
	}
	if !cfg.HasDirectChecks() || !cfg.HasProxyChecks() || !cfg.HasVPNChecks() {
		t.Fatalf("expected all groups to be present; got: direct=%d proxy=%d vpn=%d",
			len(cfg.DirectEndpoints), len(cfg.ProxyEndpoints), len(cfg.VPNEndpoints))
	}
	// Ensure HTTP-over-proxy endpoints actually carry the proxy flag/url
	foundProxyHTTP := false
	for _, ep := range cfg.ProxyEndpoints {
		if ep.TargetType == domain.TargetTypeHTTP && ep.MustUseProxy() {
			foundProxyHTTP = true
			break
		}
	}
	if !foundProxyHTTP {
		t.Fatalf("expected at least one HTTP endpoint to be marked as proxy")
	}
}
