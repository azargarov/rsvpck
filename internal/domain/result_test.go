// ./internal/domain/restult_test.go
package domain_test

import (
	"testing"
	"time"

	"github.com/azargarov/rsvpck/internal/domain"
)


func mkDNSProbe(t *testing.T) *domain.Probe {
	t.Helper()
	ep, err := domain.NewDNSEndpoint("example.com", domain.EndpointTypePublic, "dns test")
	if err != nil {
		t.Fatalf("NewDNSEndpoint error: %v", err)
	}
	return domain.NewProbe(ep)
}

func mkResultWith(t *testing.T, mode domain.ConnectivityMode, probes []domain.Probe) domain.ConnectivityResult {
	t.Helper()
	cr := domain.NewConnectivityResult(mode, probes)
	if cr.Timestamp.IsZero() {
		t.Fatalf("NewConnectivityResult must set Timestamp")
	}
	if cr.Summary == "" {
		t.Fatalf("NewConnectivityResult must set Summary")
	}
	return cr
}


func Test_NewConnectivityResult_SetsFields(t *testing.T) {
	pr := mkDNSProbe(t)
	cr := mkResultWith(t, domain.ModeDirect, []domain.Probe{*pr})

	var anyVal any = cr
	if _, ok := anyVal.(domain.ConnectivityResult); !ok {
		t.Fatalf("expected domain.ConnectivityResult, got %T", anyVal)
	}

	if cr.Mode != domain.ModeDirect {
		t.Fatalf("expected ModeDirect, got %v", cr.Mode)
	}
	if !cr.IsConnected {
		t.Fatalf("IsConnected should be true when Mode != ModeNone")
	}
	if time.Since(cr.Timestamp) < 0 || time.Since(cr.Timestamp) > time.Minute {
		t.Fatalf("Timestamp looks off: %v", cr.Timestamp)
	}
}

func Test_DetermineMode_NoSuccessfulProbes_YieldsNone(t *testing.T) {
	pr := mkDNSProbe(t)
	cr := mkResultWith(t, domain.ModeDirect, []domain.Probe{*pr})

	cr.DetermineMode()

	if cr.Mode != domain.ModeNone {
		t.Fatalf("DetermineMode() = %v, want %v", cr.Mode, domain.ModeNone)
	}
	if cr.IsConnected != cr.Mode.IsConnected() {
		t.Fatalf("IsConnected must mirror Mode.IsConnected; got IsConnected=%v, Mode=%v",
			cr.IsConnected, cr.Mode)
	}
	if cr.Summary == "" {
		t.Fatalf("Summary should be populated after DetermineMode")
	}
}

func Test_SuccessfulAndFailedProbes_MinimalCase(t *testing.T) {
	pr := mkDNSProbe(t)
	cr := mkResultWith(t, domain.ModeNone, []domain.Probe{*pr})

	if got := len(cr.SuccessfulProbes()); got != 0 {
		t.Fatalf("SuccessfulProbes() = %d, want 0", got)
	}
	if got := len(cr.FailedProbes()); got != 1 {
		t.Fatalf("FailedProbes() = %d, want 1", got)
	}
}

func Test_Summary_MatchesMode_OnConstruction(t *testing.T) {
	cases := []struct {
		mode domain.ConnectivityMode
		want string
	}{
		{domain.ModeViaVPN, "Connected via VPN."},
		{domain.ModeDirect, "Direct internet."},
		{domain.ModeViaProxy, "Internet via proxy"},
		{domain.ModeNone, "No connection"},
	}

	for _, tc := range cases {
		cr := domain.NewConnectivityResult(tc.mode, nil)
		if cr.Summary != tc.want {
			t.Fatalf("Summary for %v = %q, want %q", tc.mode, cr.Summary, tc.want)
		}
		if cr.IsConnected != (tc.mode != domain.ModeNone) {
			t.Fatalf("IsConnected for %v = %v, want %v",
				tc.mode, cr.IsConnected, tc.mode != domain.ModeNone)
		}
	}
}
