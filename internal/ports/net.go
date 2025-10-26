package ports

import (
	"context"

	"github.com/azargarov/rsvpck/internal/domain"
)


type TCPPort interface {
	CheckWithContext(ctx context.Context, ep domain.Endpoint) domain.Probe
}

type DNSPort interface {
	CheckWithContext(ctx context.Context, ep domain.Endpoint) domain.Probe
}

type HTTPPort interface {
	CheckWithContext(ctx context.Context, ep domain.Endpoint) domain.Probe
	CheckViaProxyWithContext(ctx context.Context, ep domain.Endpoint, proxyURL string) domain.Probe
}

type ICMPPort interface {
	CheckPingWithContext(ctx context.Context, ep domain.Endpoint) domain.Probe
}
