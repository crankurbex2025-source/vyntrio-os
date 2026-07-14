package backup

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func testProbePolicy(t *testing.T, opts PostStartProbePolicy) PostStartProbePolicy {
	t.Helper()
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	opts.Now = func() time.Time { return now }
	opts.Sleep = func(ctx context.Context, d time.Duration) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		now = now.Add(d)
		return nil
	}
	if opts.InitialDelay == 0 {
		opts.InitialDelay = 0
	}
	if opts.RetryInterval == 0 {
		opts.RetryInterval = time.Millisecond
	}
	if opts.Deadline == 0 {
		opts.Deadline = 5 * time.Second
	}
	return opts
}

func TestPostStartProbeTransientHealthStartupRaceSucceeds(t *testing.T) {
	policy := testProbePolicy(t, PostStartProbePolicy{
		InitialDelay:  0,
		RetryInterval: time.Millisecond,
		Deadline:      5 * time.Second,
	})
	healthCalls := 0
	readyCalls := 0
	callLog := []string{}
	err := policy.Probe(context.Background(), func(_ context.Context, path string) error {
		callLog = append(callLog, path)
		switch path {
		case "/healthz":
			healthCalls++
			if healthCalls < 3 {
				return fmt.Errorf("%w: connection refused", ErrHealthProbeFailed)
			}
			return nil
		case "/readyz":
			readyCalls++
			return nil
		default:
			t.Fatalf("unexpected path %q", path)
			return nil
		}
	})
	if err != nil {
		t.Fatalf("Probe() error: %v", err)
	}
	if healthCalls != 3 || readyCalls != 1 {
		t.Fatalf("health calls=%d ready calls=%d", healthCalls, readyCalls)
	}
	for i, path := range callLog {
		if path == "/readyz" && i < len(callLog)-1 {
			for _, later := range callLog[i+1:] {
				if later == "/healthz" {
					t.Fatalf("readyz before successful healthz settled: %v", callLog)
				}
			}
		}
	}
	lastReady := -1
	for i, path := range callLog {
		if path == "/readyz" {
			lastReady = i
		}
	}
	for i, path := range callLog {
		if path == "/healthz" && i > lastReady && lastReady >= 0 {
			t.Fatalf("healthz repeated after readyz started: %v", callLog)
		}
	}
}

func TestPostStartProbeTransientReadinessStartupRaceSucceeds(t *testing.T) {
	policy := testProbePolicy(t, PostStartProbePolicy{
		InitialDelay:  0,
		RetryInterval: time.Millisecond,
		Deadline:      5 * time.Second,
	})
	healthCalls := 0
	readyCalls := 0
	err := policy.Probe(context.Background(), func(_ context.Context, path string) error {
		switch path {
		case "/healthz":
			healthCalls++
			return nil
		case "/readyz":
			readyCalls++
			if readyCalls < 3 {
				return fmt.Errorf("%w: probe /readyz status 503", ErrReadinessProbeFailed)
			}
			return nil
		default:
			t.Fatalf("unexpected path %q", path)
			return nil
		}
	})
	if err != nil {
		t.Fatalf("Probe() error: %v", err)
	}
	if healthCalls != 1 {
		t.Fatalf("health calls = %d, want 1", healthCalls)
	}
	if readyCalls != 3 {
		t.Fatalf("ready calls = %d, want 3", readyCalls)
	}
}

func TestPostStartProbePersistentHealthFailure(t *testing.T) {
	policy := testProbePolicy(t, PostStartProbePolicy{
		InitialDelay:  0,
		RetryInterval: time.Millisecond,
		Deadline:      5 * time.Millisecond,
	})
	attempts := 0
	err := policy.Probe(context.Background(), func(_ context.Context, path string) error {
		if path == "/readyz" {
			t.Fatal("readyz probed before healthz success")
		}
		attempts++
		return fmt.Errorf("%w: connection refused", ErrHealthProbeFailed)
	})
	if !errors.Is(err, ErrHealthProbeFailed) {
		t.Fatalf("err = %v, want ErrHealthProbeFailed", err)
	}
	if attempts < 2 || attempts > 10 {
		t.Fatalf("attempts = %d, want bounded small count", attempts)
	}
}

func TestPostStartProbePersistentReadinessFailure(t *testing.T) {
	policy := testProbePolicy(t, PostStartProbePolicy{
		InitialDelay:  0,
		RetryInterval: time.Millisecond,
		Deadline:      5 * time.Millisecond,
	})
	healthCalls := 0
	readyAttempts := 0
	err := policy.Probe(context.Background(), func(_ context.Context, path string) error {
		switch path {
		case "/healthz":
			healthCalls++
			return nil
		case "/readyz":
			readyAttempts++
			return fmt.Errorf("%w: probe /readyz status 503", ErrReadinessProbeFailed)
		default:
			t.Fatalf("unexpected path %q", path)
			return nil
		}
	})
	if !errors.Is(err, ErrReadinessProbeFailed) {
		t.Fatalf("err = %v, want ErrReadinessProbeFailed", err)
	}
	if healthCalls != 1 {
		t.Fatalf("health calls = %d, want 1", healthCalls)
	}
	if readyAttempts < 2 || readyAttempts > 10 {
		t.Fatalf("ready attempts = %d, want bounded small count", readyAttempts)
	}
}

func TestPostStartProbeContextCancellation(t *testing.T) {
	policy := testProbePolicy(t, PostStartProbePolicy{
		InitialDelay:  time.Second,
		RetryInterval: time.Second,
		Deadline:      time.Minute,
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := policy.Probe(ctx, func(context.Context, string) error {
		t.Fatal("probe should not run after cancellation")
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestPostStartProbeImmediateSuccess(t *testing.T) {
	policy := testProbePolicy(t, PostStartProbePolicy{
		InitialDelay:  0,
		RetryInterval: time.Millisecond,
		Deadline:      5 * time.Second,
	})
	healthCalls := 0
	readyCalls := 0
	err := policy.Probe(context.Background(), func(_ context.Context, path string) error {
		switch path {
		case "/healthz":
			healthCalls++
		case "/readyz":
			readyCalls++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Probe() error: %v", err)
	}
	if healthCalls != 1 || readyCalls != 1 {
		t.Fatalf("health=%d ready=%d, want 1 each", healthCalls, readyCalls)
	}
}

func TestPostStartProbeErrorsDoNotLeakSensitiveContent(t *testing.T) {
	policy := testProbePolicy(t, PostStartProbePolicy{
		InitialDelay:  0,
		RetryInterval: time.Millisecond,
		Deadline:      time.Millisecond,
	})
	err := policy.Probe(context.Background(), func(_ context.Context, path string) error {
		return fmt.Errorf("%w: secret-db-content at %s", ErrHealthProbeFailed, path)
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrHealthProbeFailed) {
		t.Fatalf("err = %v, want ErrHealthProbeFailed", err)
	}
	if strings.Contains(err.Error(), "secret-db-content") {
		t.Fatalf("error leaked sensitive content: %q", err)
	}
}
