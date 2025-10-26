package icmp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/azargarov/rsvpck/internal/domain"
)

type Checker struct{}

var _ domain.ICMPChecker = (*Checker)(nil)

func (c *Checker) CheckPingWithContext(ctx context.Context, ep domain.Endpoint) domain.Probe {

	start := time.Now()
	ok, output, err := pingHostCmd(ctx, ep.Target, 1)
	latencyMs := time.Since(start).Seconds() * 1000

	if err != nil || !ok {
		detailedErr := domain.Errorf(
			domain.ErrorCodeICMPFailed,
			"Ping failed %q: %w", ep.Target, err,
		)
		return domain.NewFailedProbe(
			ep,
			mapPingError(err, ctx.Err(), output),
			detailedErr,
		)
	}

	return domain.NewSuccessfulProbe(ep, latencyMs)
}

func pingHostCmd(ctx context.Context, host string, attempts int) (bool, string, error) {
	if attempts < 1 {
		attempts = 1
	}

	var args []string
	switch runtime.GOOS {
	case "windows":
		args = []string{"-n", fmt.Sprint(attempts),"-w", "800" ,host}
	default: // linux, darwin, *bsd
		args = []string{"-c", fmt.Sprint(attempts), "-W", "1", host}
	}

	cmd := exec.CommandContext(ctx, "ping", args...)
	var out bytes.Buffer
	var errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb

	err := cmd.Run()

	output := out.String()
	if e := errb.String(); e != "" {
		output += "\n" + e
	}

	if ctx.Err() != nil {
		return false, output, ctx.Err()
	}

	//  "ping: command not found"...
	if err != nil {
	    outputLower := strings.ToLower(output)
    	if containsAny(outputLower, "invalid", "unrecognized", "illegal", "command not found") {
    	    return false, output, domain.Errorf(domain.ErrorCodeExecFailed, "ping command failed: %w (output: %q)", err, strings.TrimSpace(output))
    	}
    	return false, output, domain.Errorf(domain.ErrorCodeICMPFailed, "ping failed: %w", err)	
	}

	success := isPingSuccessful(output, runtime.GOOS, attempts)
	return success, output, nil
}

func isPingSuccessful(output, os string, attempts int) bool {
	outputLower := strings.ToLower(output)

	if os != "windows" {
		if strings.Contains(outputLower, "received") &&
			!strings.Contains(outputLower, "100% packet loss") {
			return true
		}
		if strings.Contains(outputLower, "0 received") {
			return false
		}
		return strings.Contains(outputLower, "bytes from")
	}

	if strings.Contains(outputLower, "received =") {
		return !strings.Contains(outputLower, "lost = "+fmt.Sprint(attempts))
	}
	return strings.Contains(outputLower, "reply from")
}

func mapPingError(err error, contextErr error, output string) domain.Status {

	if domain.IsErrorCode(err, domain.ErrorCodeExecFailed) {
        return domain.StatusInvalidCommand 
    }

	// 1. Context cancellation/timeout
	if contextErr != nil {
		if errors.Is(contextErr, context.DeadlineExceeded) ||
			errors.Is(contextErr, context.Canceled) {
			return domain.StatusTimeout
		}
	}

	// 2. Parse output for DNS-like errors
	outputLower := strings.ToLower(output)
	if containsAny(outputLower,
		"unknown host",
		"name or service not known",
		"nodename nor servname provided",
		"cannot resolve",
		"no address associated",
	) {
		return domain.StatusDNSFailure
	}

	// 3. Network unreachable / timeout
	if containsAny(outputLower,
		"network is unreachable",
		"host is unreachable",
		"operation timed out",
		"request timeout",
		"100% packet loss",
		"destination host unreachable",
	) {
		return domain.StatusTimeout
	}

	// 4. General failure
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
