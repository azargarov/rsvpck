package httpx

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/azargarov/rsvpck/internal/domain"
)

const singleProxyTimeout = 1 *time.Second

func GetCertificatesSmart(ctx context.Context, addr, serverName string, vpnProxy []string) ([]domain.TLSCertificate, error) {

	totalTimeout := singleProxyTimeout * time.Duration(len(vpnProxy) + 1)
    parentCtx, parentCancel := context.WithTimeout(ctx, totalTimeout)
    defer parentCancel()

    directCtx, cancel := context.WithTimeout(parentCtx, singleProxyTimeout)
    certs, err := GetCertificatesViaProxy(directCtx, addr, serverName, "")
    cancel()
    //err = errors.New("debug: force proxy fallback")
    if err != nil {
        for _, proxy := range vpnProxy {
            attemptCtx, cancel := context.WithTimeout(parentCtx, singleProxyTimeout)
            certs, proxyErr := GetCertificatesViaProxy(attemptCtx, addr, serverName, proxy)
            cancel()
            if proxyErr == nil {
                return certs, nil
            }
            //fmt.Printf("[DEBUG] proxy %s failed: %v\n", proxy, proxyErr)
        }
    }
    return certs, err  // TODO: add custom error 
}

func GetCertificatesViaProxy(ctx context.Context, targetAddr, serverName, proxyAddr string) ([]domain.TLSCertificate, error) {
	if targetAddr == "" {
		return nil, errors.New("targetAddr is required (host:port)")
	}
	if serverName == "" {
		serverName = hostPart(targetAddr)
	}

	var (
		conn net.Conn
		err  error
	)

	if proxyAddr == "" {
		conn, err = dialContext(ctx, "tcp", targetAddr)
		if err != nil {
			return nil, err
		}
		return fetchCertsOverConn(ctx, conn, serverName)
	}

	conn, err = dialThroughHTTPProxy(ctx, proxyAddr, targetAddr)
	if err != nil {
		return nil, err
	}
	return fetchCertsOverConn(ctx, conn, serverName)
}

func dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	d := &net.Dialer{}
	if deadline, ok := ctx.Deadline(); ok {
		d.Timeout = time.Until(deadline)
	} else {
		d.Timeout = 2 * time.Second
	}
	return d.DialContext(ctx, network, address)
}

func dialThroughHTTPProxy(ctx context.Context, proxyAddr, targetAddr string) (net.Conn, error) {
	u, err := parseProxyURL(proxyAddr)
	if err != nil {
		return nil, err
	}
	hostPort := u.Host
	if !strings.Contains(hostPort, ":") {
		hostPort = net.JoinHostPort(hostPort, "8080") //TODO: change default port
	}

	conn, err := dialContext(ctx, "tcp", hostPort)
	if err != nil {
		return nil, fmt.Errorf("proxy dial failed: %w", err)
	}

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	var b strings.Builder
	b.WriteString("CONNECT ")
	b.WriteString(targetAddr)
	b.WriteString(" HTTP/1.1\r\n")
	b.WriteString("Host: ")
	b.WriteString(targetAddr)
	b.WriteString("\r\n")

	if u.User != nil {
		if pw, ok := u.User.Password(); ok {
			auth := base64.StdEncoding.EncodeToString([]byte(u.User.Username() + ":" + pw))
			b.WriteString("Proxy-Authorization: Basic ")
			b.WriteString(auth)
			b.WriteString("\r\n")
		}
	}
	b.WriteString("\r\n")

	if _, err = conn.Write([]byte(b.String())); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("proxy write failed: %w", err)
	}

	br := bufio.NewReader(conn)
	statusLine, err := br.ReadString('\n')
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("proxy read failed: %w", err)
	}
	// Expect HTTP/1.1 200
	if !strings.Contains(statusLine, " 200 ") {
		var hdrs []string
		for {
			line, e := br.ReadString('\n')
			if e != nil {
				break
			}
			line = strings.TrimRight(line, "\r\n")
			if line == "" {
				break
			}
			hdrs = append(hdrs, line)
		}
		_ = conn.Close()
		return nil, fmt.Errorf("proxy CONNECT failed: %s (headers: %v)", strings.TrimSpace(statusLine), hdrs)
	}
	for {
		line, e := br.ReadString('\n')
		if e != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("proxy header read failed: %w", e)
		}
		if strings.TrimRight(line, "\r\n") == "" {
			break
		}
	}

	return conn, nil
}

func fetchCertsOverConn(ctx context.Context, rawConn net.Conn, serverName string) ([]domain.TLSCertificate, error) {
	cfg := &tls.Config{ServerName: serverName}
	tlsConn := tls.Client(rawConn, cfg)

	defer func() { _ = tlsConn.Close() }()

	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return nil, err
	}

	state := tlsConn.ConnectionState()
	now := time.Now()

	out := make([]domain.TLSCertificate, 0, len(state.PeerCertificates))
	for _, cert := range state.PeerCertificates {
		out = append(out, domain.TLSCertificate{
			Subject:   cert.Subject.String(),
			Issuer:    cert.Issuer.String(),
			NotBefore: cert.NotBefore,
			NotAfter:  cert.NotAfter,
			Valid:     !now.Before(cert.NotBefore) && !now.After(cert.NotAfter),
		})
	}
	return out, nil
}

func parseProxyURL(s string) (*url.URL, error) {
	if !strings.Contains(s, "://") {
		s = "http://" + s // assume HTTP proxy if scheme missing
	}
	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy address: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported proxy scheme %q (expected http/https)", u.Scheme)
	}
	return u, nil
}

func hostPart(addr string) string {
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}
	if i := strings.IndexByte(addr, ':'); i >= 0 {
		return addr[:i]
	}
	return addr
}
