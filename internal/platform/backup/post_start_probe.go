package backup

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// PathProbe performs one loopback GET for a fixed local endpoint path.
type PathProbe func(ctx context.Context, path string) error

// PostStartProbePolicy retries local loopback health/readiness checks after
// service restart until success or a bounded deadline expires.
type PostStartProbePolicy struct {
	InitialDelay  time.Duration
	RetryInterval time.Duration
	Deadline      time.Duration
	Sleep         func(context.Context, time.Duration) error
	Now           func() time.Time
}

func defaultPostStartProbePolicy() PostStartProbePolicy {
	return PostStartProbePolicy{
		InitialDelay:  postStartProbeInitialDelay,
		RetryInterval: postStartProbeRetryInterval,
		Deadline:      postStartProbeDeadline,
		Sleep:         probeSleep,
		Now:           time.Now,
	}
}

func (p PostStartProbePolicy) normalized() PostStartProbePolicy {
	if p.InitialDelay == 0 && p.RetryInterval == 0 && p.Deadline == 0 && p.Sleep == nil && p.Now == nil {
		return defaultPostStartProbePolicy()
	}
	out := p
	if out.RetryInterval == 0 {
		out.RetryInterval = postStartProbeRetryInterval
	}
	if out.Deadline == 0 {
		out.Deadline = postStartProbeDeadline
	}
	if out.Sleep == nil {
		out.Sleep = probeSleep
	}
	if out.Now == nil {
		out.Now = time.Now
	}
	return out
}

// Probe waits for /healthz then /readyz success on the fixed loopback base URL.
func (p PostStartProbePolicy) Probe(ctx context.Context, probe PathProbe) error {
	policy := p.normalized()
	deadline := policy.deadlineTime(ctx)
	if err := policy.probeUntilSuccess(ctx, deadline, "/healthz", ErrHealthProbeFailed, probe); err != nil {
		return err
	}
	return policy.probeUntilSuccess(ctx, deadline, "/readyz", ErrReadinessProbeFailed, probe)
}

func (p PostStartProbePolicy) deadlineTime(ctx context.Context) time.Time {
	policyDeadline := p.Now().Add(p.Deadline)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(policyDeadline) {
		return ctxDeadline
	}
	return policyDeadline
}

func (p PostStartProbePolicy) probeUntilSuccess(ctx context.Context, deadline time.Time, path string, sentinel error, probe PathProbe) error {
	if err := p.Sleep(ctx, p.InitialDelay); err != nil {
		return probeContextFailure(err, sentinel)
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !p.Now().Before(deadline) {
			return sentinel
		}
		if err := probe(ctx, path); err == nil {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if !p.Now().Before(deadline) {
			return sentinel
		}
		if err := p.Sleep(ctx, p.RetryInterval); err != nil {
			return probeContextFailure(err, sentinel)
		}
	}
}

func probeContextFailure(err, sentinel error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return fmt.Errorf("%w: %v", sentinel, err)
}

func probeSleep(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
