package tcp

import (
	"context"
	"errors"
	"github.com/azargarov/rsvpck/internal/domain"
	"net"
	"strings"
	"syscall"
	"time"
)

const localTimeOut = 1 * time.Second

type Checker struct{}

// CheckWithContext executes TCP-connect to the target  "host:port"
func (c Checker) CheckWithContext(ctx context.Context, ep domain.Endpoint) domain.Probe {
	if _, _, err := net.SplitHostPort(ep.Target); err != nil {
		return domain.NewFailedProbe(
			ep,
			domain.StatusInvalid,
			errors.New("invalid target format, expected host:port"),
		)
	}

	start := time.Now()

	dialer := &net.Dialer{
		Timeout:   localTimeOut,
		KeepAlive: 0,
	}

	conn, err := dialer.DialContext(ctx, "tcp", ep.Target)
	latencyMs := time.Since(start).Seconds() * 1000

	if err != nil {
		status := mapErrorToStatus(err, ctx.Err())
		return domain.NewFailedProbe(
			ep,
			status,
			err,
		)
	}
	conn.Close()

	return domain.NewSuccessfulProbe(
		ep,
		latencyMs,
	)
}

func mapErrorToStatus(err, contextErr error) domain.Status {
	if contextErr != nil {
		if errors.Is(contextErr, context.DeadlineExceeded) ||
			errors.Is(contextErr, context.Canceled) {
			return domain.StatusTimeout
		}
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return domain.StatusTimeout
		}
	}

	if opErr, ok := err.(*net.OpError); ok {
		if syscallErr, ok := opErr.Err.(*syscall.Errno); ok {
			switch *syscallErr {
			case syscall.ECONNREFUSED:
				return domain.StatusConnectionRefused
			case syscall.ENETUNREACH, syscall.EHOSTUNREACH:
				return domain.StatusTimeout
			}
		}
	}

	errStr := err.Error()
	if containsAny(errStr, "connection refused", "connection reset") {
		return domain.StatusConnectionRefused
	}
	if containsAny(errStr, "timeout", "i/o timeout", "context deadline exceeded") {
		return domain.StatusTimeout
	}

	return domain.StatusInvalid
}

func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if len(substr) > 0 && (len(s) >= len(substr) && containsIgnoreCase(s, substr)) {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return len(s) >= len(substr) && (len(substr) == 0 || strings.Contains(s, substr))
}
