package netadapter

import (
	"context"
	"fmt"
	"github.com/azargarov/rsvpck/internal/domain"
	"net"
	"time"
)

const(
	dialTimeOut = 1*time.Second
)

type TCPDialer struct{}

var _ domain.TCPChecker = (*TCPDialer)(nil)

func (d *TCPDialer) CheckWithContext(ctx context.Context, ep domain.Endpoint) domain.Probe {
	addr := ep.Target
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, dialTimeOut)
	latencyMs := time.Since(start).Seconds() * 1000

	if err != nil {
		return domain.NewFailedProbe(
			ep,
			domain.StatusConnectionRefused,
			err,
		)
	}
	conn.Close()

	return domain.NewSuccessfulProbe(ep, latencyMs)
}

func (d *TCPDialer) HTTPDo(ctx context.Context, req *domain.Request) (*domain.Response, time.Duration, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (d *TCPDialer) ICMPPing(ctx context.Context, host string, count int) (time.Duration, error) {
	return 0, fmt.Errorf("not implemented")
}
