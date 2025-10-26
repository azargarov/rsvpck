
package app

import (
	"context"

	"github.com/azargarov/rsvpck/internal/domain"
	"github.com/azargarov/rsvpck/internal/ports"
)

type PortProber interface {
	Run(ctx context.Context, ep domain.Endpoint) domain.Probe
}

type CompositeProber struct {
	tcp  ports.TCPPort
	dns  ports.DNSPort
	http ports.HTTPPort
	icmp ports.ICMPPort
}

func NewCompositeProber(
	tcp  ports.TCPPort,
	dns  ports.DNSPort,
	http ports.HTTPPort,
	icmp ports.ICMPPort,
) *CompositeProber {
	return &CompositeProber{tcp: tcp, dns: dns, http: http, icmp: icmp}
}

func (p *CompositeProber) Run(ctx context.Context, ep domain.Endpoint) domain.Probe {
	switch ep.GetTargetType() {
	case domain.TargetTypeICMP:
		return p.icmp.CheckPingWithContext(ctx, ep)
	case domain.TargetTypeTCP:
		return p.tcp.CheckWithContext(ctx, ep)
	case domain.TargetTypeHTTP:
		if ep.MustUseProxy() {
			return p.http.CheckViaProxyWithContext(ctx, ep, ep.Proxy.URL())
		}
		return p.http.CheckWithContext(ctx, ep)
	case domain.TargetTypeDNS:
		return p.dns.CheckWithContext(ctx, ep)
	default:
		return domain.NewFailedProbe(ep, domain.StatusInvalid,
			domain.Errorf(domain.ErrorCodeInvalidConfig, "unknown target type"))
	}
}
