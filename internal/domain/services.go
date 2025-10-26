package domain

import (
	"context"
	"io"
)

type TCPChecker interface {
	CheckWithContext(ctx context.Context, ep Endpoint) Probe
}

type DNSChecker interface {
	CheckWithContext(ctx context.Context, ep Endpoint) Probe
}

type HTTPChecker interface {
	CheckWithContext(ctx context.Context, ep Endpoint) Probe
	CheckViaProxyWithContext(ctx context.Context, ep Endpoint, proxyURL string) Probe
}

type ICMPChecker interface {
	CheckPingWithContext(ctx context.Context, ep Endpoint) Probe
}

type HostChecker interface {
	GetCRMInfo(ctx context.Context) HostInfo
}

type Renderer interface {
	Render(w io.Writer, result ConnectivityResult) error
}

type StringRenderer interface {
	Render(w io.Writer, name, str string) error
	RenderArray(w io.Writer, name string, data []any) error
}