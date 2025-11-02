package domain_test

import (
	"testing"
	"time"

	"github.com/azargarov/rsvpck/internal/domain"
)

func mkProbe(ep domain.Endpoint, st domain.Status, lat float64) domain.Probe {
	return domain.Probe{
		Endpoint:  ep,
		Status:    st,
		LatencyMs: lat,
		Timestamp: time.Now(),
	}
}

func TestDetermineMode_Direct(t *testing.T) {
	dns := domain.MustNewDNSEndpoint("example.com", domain.EndpointTypePublic, "")
	tcp := domain.MustNewTCPEndpoint("example.com:443", domain.EndpointTypePublic, "")

	r := domain.ConnectivityResult{
		Probes: []domain.Probe{
			mkProbe(dns, domain.StatusPass, 1),
			mkProbe(tcp, domain.StatusPass, 5),
		},
	}
	r.DetermineMode()

	if r.Mode != domain.ModeDirect || !r.IsConnected {
		t.Fatalf("want ModeDirect connected, got %v connected=%v", r.Mode, r.IsConnected)
	}
}

//TODO: find proxy to test
func TestDetermineMode_ViaProxy(t *testing.T) {
	httpEp := domain.MustNewHTTPEndpoint("https://example.com", domain.EndpointTypePublic, false,"", "")
	httpEp.SetProxy("http://proxy.local:8080")

	r := domain.ConnectivityResult{
		Probes: []domain.Probe{
			mkProbe(httpEp, domain.StatusPass, 12),
		},
	}
	r.DetermineMode()

	if r.Mode != domain.ModeViaProxy || !r.IsConnected {
		t.Fatalf("want ModeViaProxy connected, got %v connected=%v", r.Mode, r.IsConnected)
	}
}

func TestDetermineMode_VPN(t *testing.T) {
	vpnTCP := domain.MustNewTCPEndpoint("10.0.0.5:443", domain.EndpointTypeVPN, "vpn")
	r := domain.ConnectivityResult{
		Probes: []domain.Probe{
			mkProbe(vpnTCP, domain.StatusPass, 3),
		},
	}
	r.DetermineMode()

	if r.Mode != domain.ModeViaVPN || !r.IsConnected {
		t.Fatalf("want ModeViaVPN connected, got %v connected=%v", r.Mode, r.IsConnected)
	}
}

func TestDetermineMode_None(t *testing.T) {
	dns := domain.MustNewDNSEndpoint("nosuch.example.invalid", domain.EndpointTypePublic, "")
	r := domain.ConnectivityResult{
		Probes: []domain.Probe{
			mkProbe(dns, domain.StatusFail, 0),
		},
	}
	r.DetermineMode()

	if r.Mode != domain.ModeNone || r.IsConnected {
		t.Fatalf("want ModeNone disconnected, got %v connected=%v", r.Mode, r.IsConnected)
	}
}

func TestEndpoints_Validation(t *testing.T) {
	if _, err := domain.NewHTTPEndpoint("not-a-url", domain.EndpointTypePublic, ""); err == nil {
		t.Fatalf("expected HTTP endpoint validation error")
	}
	if _, err := domain.NewTCPEndpoint("noport", domain.EndpointTypePublic, ""); err == nil {
		t.Fatalf("expected TCP endpoint validation error")
	}
	if _, err := domain.NewICMPEndpoint("", domain.EndpointTypePublic, ""); err == nil {
		t.Fatalf("expected ICMP endpoint validation error")
	}
	if _, err := domain.NewDNSEndpoint("http://bad", domain.EndpointTypePublic, ""); err == nil {
		t.Fatalf("expected DNS endpoint validation error")
	}
}
