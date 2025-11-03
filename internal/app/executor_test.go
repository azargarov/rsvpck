package app

import (
	"context"
	//"sync/atomic"
	"sync"
	"testing"
	"time"
	wp "github.com/azargarov/go-utils/wpool"
	"github.com/azargarov/rsvpck/internal/domain"
)


type fakeProber struct {
    calls map[string]int32
    mu    sync.Mutex
}

func newFakeProber() *fakeProber {
    return &fakeProber{
        calls: make(map[string]int32),
    }
}

func (p *fakeProber) Run(ctx context.Context, ep domain.Endpoint) domain.Probe {
    key := ep.String() 

    p.mu.Lock()
    p.calls[key]++
    n := p.calls[key]
    p.mu.Unlock()

    // first call for THIS endpoint → fail
    if n == 1 {
        return domain.NewFailedProbe(ep, domain.StatusTimeout,
            domain.Errorf(domain.ErrorCodeTCPTimedOut, "simulated"))
    }

    // second (and later) call for THIS endpoint → success
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
	p := newFakeProber()
	// small pool to exercise concurrency
	rp := *wp.GetDefaultRP()
	rp.Attempts = 2
	w := wp.NewPool[probeJob](2, rp)

	ex := NewExecutorWithPool(p, domain.PolicyExhaustive, w)

	cfg, err := domain.NewNetTestConfig(
		[]domain.Endpoint{domain.MustNewTCPEndpoint("google.com:443", domain.EndpointTypeVPN, "vpn")},
		[]domain.Endpoint{
			domain.MustNewDNSEndpoint("google.com", domain.EndpointTypePublic, "dns"),
			domain.MustNewTCPEndpoint("google.com:443", domain.EndpointTypePublic, "tcp"),
			mkHTTP("https://google.com", false),
		},
		[]domain.Endpoint{
			mkHTTP("https://example.com", false),
		},
		"",
		nil,
	)
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res := ex.Run(ctx, cfg)

	for i, pr := range res.Probes {
		if !pr.IsSuccessful() {
			t.Fatalf("probe %d not successful: %+v", i, pr)
		}
	}
}
