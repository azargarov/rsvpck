package dns

import (
	"context"
	"errors"
	"github.com/azargarov/rsvpck/internal/domain"
	"net"
	"net/netip"
	"strings"
	"time"
)

const dnsTimeout = 800 * time.Millisecond

type Checker struct{}

var _ domain.DNSChecker = (*Checker)(nil)

func (r Checker) CheckWithContext(parentCtx context.Context, ep domain.Endpoint) domain.Probe {
	ctx, cancel := context.WithTimeout(parentCtx, dnsTimeout)
    defer cancel()

	start := time.Now()
	_, err := net.DefaultResolver.LookupHost(ctx, ep.Target)
	latencyMs := time.Since(start).Seconds() * 1000

	if err != nil {
		detailedErr := domain.Errorf(
			domain.ErrorCodeDNSUnresolvable,
			"DNS resolution failed for %q: %w", ep.Target, err,
		)
		return domain.NewFailedProbe(
			ep,
			mapDNSError(err, ctx.Err()),
			detailedErr,
		)
	}
	return domain.NewSuccessfulProbe(ep, latencyMs)
}

func (r *Checker) LookupHost(ctx context.Context, host string, timeout time.Duration) (ips []netip.Addr, err error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var dialer = &net.Dialer{Timeout: 800 * time.Millisecond}

	var fastResolver = &net.Resolver{
    	PreferGo: true,
    	Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
    	    // Force a short DNS timeout regardless of system resolver settings
    	    return dialer.DialContext(ctx, "udp", "8.8.8.8:53")
    	},
	}
	strIPs, err := fastResolver.LookupHost(ctx, host)
	//strIPs, err := net.DefaultResolver.LookupHost(ctx, host)
	if err != nil {
		return nil, err
	}
	addrs := make([]netip.Addr, 0, len(strIPs))
	for _, ip := range strIPs {
		if addr, perr := netip.ParseAddr(ip); perr == nil {
			addrs = append(addrs, addr)
		}
	}

	return addrs, nil
}

func mapDNSError(err, contextErr error) domain.Status {
	if contextErr != nil {
		if errors.Is(contextErr, context.DeadlineExceeded) ||
			errors.Is(contextErr, context.Canceled) {
			return domain.StatusTimeout
		}
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		if dnsErr.Timeout() {
			return domain.StatusTimeout
		}
		if dnsErr.IsNotFound {
			return domain.StatusDNSFailure
		}
		if dnsErr.Err == "no such host" || dnsErr.Err == "server misbehaving" {
			return domain.StatusDNSFailure
		}
	}

	errStr := err.Error()
	if containsAny(errStr,
		"no such host",
		"server misbehaving",
		"cannot unmarshal DNS",
		"host not found",
		"NXDOMAIN",
	) {
		return domain.StatusDNSFailure
	}

	if containsAny(errStr,
		"timeout",
		"i/o timeout",
		"network is unreachable",
		"connection refused",
	) {
		return domain.StatusTimeout
	}

	return domain.StatusInvalid
}

func containsAny(s string, substrs ...string) bool {
	sLower := strings.ToLower(s)
	for _, substr := range substrs {
		if substr != "" && strings.Contains(sLower, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}
