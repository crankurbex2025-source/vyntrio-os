package backup

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// SystemdController controls a systemd unit.
type SystemdController struct {
	UnitName string
	Timeout  time.Duration
}

func (c SystemdController) unit() string {
	if c.UnitName != "" {
		return c.UnitName
	}
	return ServiceName
}

func (c SystemdController) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return defaultServiceTimeout
}

func (c SystemdController) IsActive(ctx context.Context) (bool, error) {
	out, err := c.run(ctx, "is-active", c.unit())
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "active", nil
}

func (c SystemdController) Stop(ctx context.Context) error {
	_, err := c.run(ctx, "stop", c.unit())
	return err
}

func (c SystemdController) IsInactive(ctx context.Context) (bool, error) {
	out, err := c.runIsActive(ctx, c.unit())
	if err != nil {
		return false, err
	}
	state := strings.TrimSpace(string(out))
	return state == "inactive" || state == "failed", nil
}

func (c SystemdController) Start(ctx context.Context) error {
	_, err := c.run(ctx, "start", c.unit())
	return err
}

func (c SystemdController) runIsActive(ctx context.Context, unit string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout())
	defer cancel()
	cmd := exec.CommandContext(ctx, "systemctl", "is-active", unit)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return out, nil
	}
	// systemctl is-active uses a non-zero exit status for non-active units but
	// still prints the state on stdout.
	if len(out) > 0 {
		return out, nil
	}
	return nil, fmt.Errorf("systemctl is-active %s: %w", unit, err)
}

func (c SystemdController) run(ctx context.Context, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout())
	defer cancel()
	cmd := exec.CommandContext(ctx, "systemctl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("systemctl %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}

// LoopbackHealthProber checks local health endpoints with bounded post-start retries.
type LoopbackHealthProber struct {
	Client  *http.Client
	BaseURL string
	Policy  PostStartProbePolicy
}

func (p LoopbackHealthProber) Probe(ctx context.Context) error {
	return p.Policy.normalized().Probe(ctx, func(ctx context.Context, path string) error {
		sentinel := ErrHealthProbeFailed
		if path == "/readyz" {
			sentinel = ErrReadinessProbeFailed
		}
		return p.probePath(ctx, path, sentinel)
	})
}

func (p LoopbackHealthProber) probePath(ctx context.Context, path string, sentinel error) error {
	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: defaultHTTPTimeout}
	}
	base := p.BaseURL
	if base == "" {
		base = HealthBaseURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", sentinel, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", sentinel, err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: probe %s status %d", sentinel, path, resp.StatusCode)
	}
	return nil
}

// OSRootChecker reports whether the effective UID is zero.
type OSRootChecker struct{}

func (OSRootChecker) IsRoot() bool {
	return geteuid() == 0
}

var geteuid = syscall.Geteuid
