package domain

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

type EndpointTargetType int

const (
	TargetTypeHTTP EndpointTargetType = iota // http:// or https://
	TargetTypeTCP                            // host:port for TCP-connect
	TargetTypeICMP                           // to ping
	TargetTypeDNS
)

func (t EndpointTargetType) String() string {
	switch t {
	case TargetTypeHTTP:
		return "http"
	case TargetTypeTCP:
		return "tcp"
	case TargetTypeDNS:
		return "DNS"
	case TargetTypeICMP:
		return "icmp"
	default:
		return "unknown"
	}
}

type EndpointType int

const (
	EndpointTypePublic EndpointType = iota
	EndpointTypeVPN
)

func (t EndpointType) String() string {
	switch t {
	case EndpointTypePublic:
		return "public"
	case EndpointTypeVPN:
		return "vpn"
	default:
		return "unknown"
	}
}

type Endpoint struct {
	Target        string 				// "http://1.1.1.1" or "10.10.0.1:443"
	TargetType    EndpointTargetType
	Type          EndpointType
	Proxy         ProxyConfig
	Description   string
}

func (e Endpoint) MustUseProxy() bool {
	return e.TargetType == TargetTypeHTTP && e.Proxy.MustUseProxy()
}

func (e *Endpoint) SetProxy(proxy string){
	e.Proxy.Set(proxy)
}

func (e Endpoint) String() string {
	str := fmt.Sprintf("Target: %s, TType: %s, Type: %s, Descr: %s",
		e.Target, e.TargetType.String(), e.Type.String(), e.Description)
	return str
}

func (e Endpoint) Key() string {
    return fmt.Sprintf("%s|%s|%s|%t|%s",
        e.TargetType, e.Type, e.Target, e.MustUseProxy(), e.Proxy.URL())
}

func NewHTTPEndpoint(url string, typ EndpointType, description string) (Endpoint, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return Endpoint{}, errors.New("HTTP endpoint must start with http:// or https://")
	}
	return Endpoint{
		Target:      url,
		TargetType:  TargetTypeHTTP,
		Type:        typ,
		Description: description,
	}, nil
}

func NewTCPEndpoint(hostPort string, typ EndpointType, description string) (Endpoint, error) {
	_, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		return Endpoint{}, fmt.Errorf("invalid host:port format: %w", err)
	}
	return Endpoint{
		Target:      hostPort,
		TargetType:  TargetTypeTCP,
		Type:        typ,
		Description: description,
	}, nil
}

func NewICMPEndpoint(host string, typ EndpointType, description string) (Endpoint, error) {
	if strings.TrimSpace(host) == "" {
		return Endpoint{}, errors.New("ICMP host cannot be empty")
	}

	if strings.Contains(host, ":") && !strings.Contains(host, "]") {
		if !strings.HasPrefix(host, "[") || !strings.HasSuffix(host, "]") {
			return Endpoint{}, errors.New("ICMP host must not contain a port (no colon allowed)")
		}
	}

	return Endpoint{
		Target:      strings.TrimSpace(host),
		TargetType:  TargetTypeICMP,
		Type:        typ,
		Description: description,
	}, nil
}

func NewDNSEndpoint(host string, typ EndpointType, description string) (Endpoint, error) {
	if strings.TrimSpace(host) == "" {
		return Endpoint{}, errors.New("DNS host cannot be empty")
	}
	// Optional: basic validation (no scheme, no port)
	if strings.Contains(host, "://") || strings.Contains(host, ":") && !strings.Contains(host, "]") {
		return Endpoint{}, errors.New("DNS endpoint must be a plain hostname (no URL, no port)")
	}
	return Endpoint{
		Target:      strings.TrimSpace(host),
		TargetType:  TargetTypeDNS,
		Type:        typ,
		Description: description,
	}, nil
}

func (e Endpoint) GetTargetType() EndpointTargetType {
	return e.TargetType
}

func (e Endpoint) IsVPN() bool {
	return e.Type == EndpointTypeVPN
}

func (e Endpoint) IsPublic() bool {
	return e.Type == EndpointTypePublic
}

func (e Endpoint) IsDirectType() bool {
	return (e.TargetType == TargetTypeTCP || e.TargetType == TargetTypeHTTP) && !e.MustUseProxy() 
}

func (e Endpoint) IsProxyType() bool {
	return (e.TargetType == TargetTypeTCP || e.TargetType == TargetTypeHTTP) && e.MustUseProxy() 
}

func (e Endpoint) IsICMP() bool {
	return e.TargetType == TargetTypeICMP
}

func (e Endpoint) IsDNS() bool {
	return e.TargetType == TargetTypeDNS
}

func (e Endpoint) IsTCP() bool {
	return e.TargetType == TargetTypeTCP
}

func (e Endpoint) IsHTTP() bool {
	return e.TargetType == TargetTypeHTTP
}
