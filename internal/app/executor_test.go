package app

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	wp "github.com/azargarov/go-utils/wpool"
	"github.com/azargarov/rsvpck/internal/domain"
)

type fakeProber struct {
	call int32
}

func (p *fakeProber) Run(ctx context.Context, ep domain.Endpoint) domain.Probe {
	n := atomic.AddInt32(&p.call, 1)
	// first call for each endpoint key: timeout (retryable)
	// next call: success
	if n%2 == 1 {
		return domain.NewFailedProbe(ep, domain.StatusTimeout, domain.Errorf(domain.ErrorCodeTCPTimedOut, "simulated"))
	}
	return domain.NewSuccessfulProbe(ep, 1.23)
}

func mkHTTP(target string, overProxy bool) domain.Endpoint {
	ep := domain.MustNewHTTPEndpoint(target, domain.EndpointTypePublic, false,"", "http")
	if overProxy {
		ep.SetProxy("http://proxy.local:8080")
	}
	return ep
}

func TestExecutor_Run_AllProbesCompletedWithRetries(t *testing.T) {
		p := &fakeProber{}
	// small pool to exercise concurrency
	rp := *wp.GetDefaultRP()
	rp.Attempts = 2
	w := wp.NewPool[probeJob](4, rp)

	ex := NewExecutorWithPool(p, domain.PolicyExhaustive, w)

	cfg, err := domain.NewNetTestConfig(
		[]domain.Endpoint{domain.MustNewTCPEndpoint("10.0.0.1:443", domain.EndpointTypeVPN, "vpn")},
		[]domain.Endpoint{
			domain.MustNewDNSEndpoint("example.com", domain.EndpointTypePublic, "dns"),
			domain.MustNewTCPEndpoint("example.com:443", domain.EndpointTypePublic, "tcp"),
			mkHTTP("https://example.com", false),
		},
		[]domain.Endpoint{
			mkHTTP("https://example.com", true),
		},
		"http://proxy.local:8080",
		nil,
	)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res := ex.Run(ctx, cfg)

	if len(res.Probes) != 1+3+1 {
		t.Fatalf("unexpected probes count: %d", len(res.Probes))
	}
	// After one retry, all should become StatusPass given fakeProber logic
	for i, pr := range res.Probes {
		if !pr.IsSuccessful() {
			t.Fatalf("probe %d not successful: %+v", i, pr)
		}
	}
}
