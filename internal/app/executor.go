package app

import (
	"context"
	"fmt"
	"time"

	wp "github.com/azargarov/go-utils/wpool"
	lg "github.com/azargarov/go-utils/zlog"
	"github.com/azargarov/rsvpck/internal/domain"
)
const (
	attempts = 2
	initialTimeout = 200 * time.Millisecond
	maxTimeout = 2 * time.Second
	totalMaxWorkers = 25
)


type Executor struct {
	policy      domain.ExecutionPolicy
	prober		PortProber
	pool 		*wp.Pool[probeJob]
}

type probeJob struct{
	Index 	int
	Ep 		domain.Endpoint
}

func NewExecutorWithPool(prober PortProber, policy domain.ExecutionPolicy, pool *wp.Pool[probeJob]) *Executor {
	return &Executor{prober: prober, policy: policy, pool: pool}
}

func NewExecutor(prober PortProber, policy domain.ExecutionPolicy) *Executor {
	w := wp.NewPool[probeJob](totalMaxWorkers, *wp.GetDefaultRP())
    return NewExecutorWithPool(prober, policy, w)
}

func (e *Executor) Run(ctx context.Context, config domain.NetTestConfig) domain.ConnectivityResult {
	
	defer e.pool.Stop()

	all := make([]domain.Endpoint, 0, len(config.DirectEndpoints)+len(config.ProxyEndpoints)+len(config.VPNEndpoints))
	all = append(all, config.DirectEndpoints...)
    all = append(all, config.ProxyEndpoints...)
    all = append(all, config.VPNEndpoints...)
 	probes := e.runEndpointCheck(ctx, all)

	return domain.AnalyzeConnectivity(probes, config)
}

func (e Executor) runEndpointCheck(ctx context.Context, endpoints []domain.Endpoint) []domain.Probe {

	n := len(endpoints)
	if n == 0 {
		return nil
	}

	// pass logger to pool to discard messages
	logger := lg.NewDiscard()
    ctx = lg.Attach(ctx, logger)
	
	results := make([]domain.Probe, n)
	done := make(chan struct{}, n)

	for i, ep := range endpoints {
		job := wp.Job[probeJob]{
			Payload: probeJob{Index: i, Ep: ep},
			Ctx:     ctx,
			Fn: func(pj probeJob) error {
				probe := e.prober.Run(ctx, pj.Ep) 
				results[pj.Index] = probe

				if !probe.IsSuccessful() && (probe.Status == domain.StatusTimeout) {
					return fmt.Errorf("retryable: timeout")
				}
				return nil
			},
			CleanupFunc: func() { done <- struct{}{} },
			Retry: &wp.RetryPolicy{Attempts: attempts, Initial: initialTimeout, Max: maxTimeout},
		}
		_ = e.pool.Submit(job)
	}

	// Wait for all
	for range n {
		select {
		case <-done:
		case <-ctx.Done():
			return results
		}
	}
	return results
}
