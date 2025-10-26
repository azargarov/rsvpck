package domain

import (
	"fmt"
	"time"
)

type Probe struct {
	Endpoint  Endpoint
	Status    Status
	LatencyMs float64
	Error     string
	Timestamp time.Time
}

func (p Probe) IsSuccessful() bool {
	return p.Status == StatusPass
}

func (p Probe) IsSkipped() bool {
	return p.Status == StatusSkipped
}

func (p Probe) IsVPNProbe() bool {
	return p.Endpoint.IsVPN()
}

func (p Probe) IsDNSProbe() bool {
	return p.Endpoint.IsDNS()
}

func (p Probe) String() string {
	var str string
	if !p.IsSuccessful() {
		str = fmt.Sprintf("Endpoint: %s, Status: %s, Error: %s, ",
			p.Endpoint.String(), p.Status.String(), p.Error)
		return str

	}
	str = fmt.Sprintf("Endpoint: %v, Status: %s, Latency: %0.2f, ",
		p.Endpoint, p.Status.String(), p.LatencyMs)

	return str
}

func NewSuccessfulProbe(endpoint Endpoint, latencyMs float64) Probe {
	p := NewProbe(endpoint)
	p.MarkSuccess(latencyMs)
	return *p 
}
func NewFailedProbe(endpoint Endpoint, status Status, err error) Probe {
	p := NewProbe(endpoint)
	p.MarkFailure(status, err)
	return *p 
}

func NewProbe(endpoint Endpoint) *Probe{
	return &Probe{Endpoint: endpoint}
}

func (p *Probe) MarkSuccess(latMs float64) {
	p.Status = StatusPass
	p.LatencyMs = latMs
	p.Error = ""
	p.Timestamp = time.Now()
}

func (p *Probe) MarkFailure(st Status, err error) {
	p.Status = st
	p.Error = safeErr(err)
	p.Timestamp = time.Now()
}

func safeErr(err error) string {
	if err == nil { return "" }
	return err.Error()
}