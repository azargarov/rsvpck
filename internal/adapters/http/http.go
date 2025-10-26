package http

import (
	"context"
	"errors"
	"github.com/azargarov/rsvpck/internal/domain"
	"net/http"
	"net/url"
	//"net"
	"time"
)

const (
	requestTimeOut = 1 * time.Second
)

type Checker struct{}

var _ domain.HTTPChecker = (*Checker)(nil)

func (c Checker) CheckWithContext(ctx context.Context, ep domain.Endpoint) domain.Probe {
	return c.doRequest(ctx, ep, nil)
}

func (c Checker) CheckViaProxyWithContext(ctx context.Context, ep domain.Endpoint, proxyURL string) domain.Probe {
	proxyParsed, err := url.Parse(proxyURL)
	if err != nil {
		return domain.NewFailedProbe(
			ep,
			domain.StatusInvalid,
			errors.New("invalid proxy URL: "+err.Error()),
		)
	}
	return c.doRequest(ctx, ep, proxyParsed)
}

// TODO: test with proxy
// doRequest is the shared HTTP execution logic.
func (c Checker) doRequest(ctx context.Context, ep domain.Endpoint, proxyURL *url.URL) domain.Probe {

	//proxyURL = nil
	var transport http.RoundTripper = http.DefaultTransport
	if proxyURL != nil {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.Proxy = func(*http.Request) (*url.URL, error) {
			return proxyURL, nil
		}
		transport = t
	}

	//var dialer = &net.Dialer{Timeout: requestTimeOut}
	//var fastResolver = &net.Resolver{
	//    PreferGo: true,
	//    Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
	//        return dialer.DialContext(ctx, "udp", "8.8.8.8:53")
	//    },
	//}
	////t := http.DefaultTransport.(*http.Transport).Clone()
	//t := transport.(*http.Transport).Clone()
	//t.DialContext = (&net.Dialer{Timeout: 500 * time.Millisecond}).DialContext
	//t.TLSHandshakeTimeout = 500 * time.Millisecond
	//t.ResponseHeaderTimeout = 500 * time.Millisecond
	//net.DefaultResolver = fastResolver  
	//client := &http.Client{ Transport: t, Timeout: 1 * time.Second }



	client := &http.Client{
		Transport: transport,
		Timeout:   requestTimeOut, 
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // don't follow redirects
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", ep.Target, nil)
	if err != nil {
		info := classifyHTTPError(err, ctx.Err())
		detailedErr := domain.Errorf(
			info.ErrorCode,
			"HTTP test failed for %q: %w", ep.Target, err,
		)
		return domain.NewFailedProbe(
			ep,
			domain.StatusInvalid,
			detailedErr,
		)
	}

	// a user-agent to avoid 403s
	req.Header.Set("User-Agent", "rsvpck/0.2 (network tester)")

	start := time.Now()
	resp, err := client.Do(req)
	latencyMs := time.Since(start).Seconds() * 1000

	if err != nil {
		info := classifyHTTPError(err, ctx.Err())
		detailedErr := domain.Errorf(
			info.ErrorCode,
			"HTTP test failed for %q: %w", ep.Target, err,
		)
		return domain.NewFailedProbe(
			ep,
			info.Status,
			detailedErr,
		)
	}
	defer resp.Body.Close()

	// TODO: Consider 2xx and 3xx (redirects) as success for connectivity
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return domain.NewSuccessfulProbe(
			ep,
			latencyMs,
		)
	}
	var errorCode domain.ErrorCode
	switch {
	case resp.StatusCode == 407:
		errorCode = domain.ErrorCodeProxyAuthRequired
	case resp.StatusCode >= 500:
		errorCode = domain.ErrorCodeHTTPBadStatus
	case resp.StatusCode >= 400:
		errorCode = domain.ErrorCodeHTTPClientError
	default:
		errorCode = domain.ErrorCodeHTTPBadStatus
	}
	detailedErr := domain.Errorf(
		errorCode,
		"HTTP request to %q returned %s", ep.Target, resp.Status,
	)
	return domain.NewFailedProbe(
		ep,
		domain.StatusHTTPError,
		detailedErr,
	)
}
