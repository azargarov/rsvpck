package httpx

import "testing"

func TestHostPart(t *testing.T) {
	cases := map[string]string{
		"example.com:443": "example.com",
		"example.com":     "example.com",
		"[2001:db8::1]:443": "2001:db8::1",
	}
	for in, want := range cases {
		got := hostPart(in)
		if got != want {
			t.Fatalf("hostPart(%q)=%q; want %q", in, got, want)
		}
	}
}

func TestParseProxyURL(t *testing.T) {
	ok := []string{
		"http://proxy.local:8080",
		"https://proxy.local:8080",
		"proxy.local:8080", // scheme-less should default to http
	}
	for _, s := range ok {
		if _, err := parseProxyURL(s); err != nil {
			t.Fatalf("parseProxyURL(%q) unexpected error: %v", s, err)
		}
	}
	bad := []string{
		"socks5://proxy.local:1080",
		"://broken",
	}
	for _, s := range bad {
		if _, err := parseProxyURL(s); err == nil {
			t.Fatalf("parseProxyURL(%q) expected error", s)
		}
	}
}
