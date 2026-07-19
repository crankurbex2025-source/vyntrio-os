package installpreflight

import (
	"context"
	"fmt"
	"strings"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/releaseartifact"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/storageinventory"
)

// Checker performs read-only install target and media preflight.
type Checker struct {
	Collector storageinventory.Collector
	Verifier  releaseartifact.Verifier
}

// NewChecker creates a preflight checker with default platform collectors.
func NewChecker(stateDir string) Checker {
	if strings.TrimSpace(stateDir) == "" {
		stateDir = DefaultStateDir
	}
	return Checker{
		Collector: storageinventory.NewCollector(stateDir, storageinventory.CollectorDeps{}),
		Verifier:  releaseartifact.NewVerifier(),
	}
}

// Run executes target and optional media preflight checks.
func (c Checker) Run(ctx context.Context, target TargetRequest, media MediaRequest) (Result, error) {
	_ = ctx
	result := Result{}

	targetResult, targetErr := c.checkTarget(target)
	result.Target = targetResult

	mediaResult, mediaErr := c.checkMedia(media)
	result.Media = mediaResult

	switch {
	case targetErr != nil && mediaErr != nil:
		return result, fmt.Errorf("%w", ErrPreflightFailed)
	case targetErr != nil:
		return result, targetErr
	case mediaErr != nil:
		return result, mediaErr
	default:
		return result, nil
	}
}

func (c Checker) checkTarget(req TargetRequest) (TargetResult, error) {
	diskID := strings.TrimSpace(req.DiskID)
	if diskID == "" {
		return TargetResult{
			Status:  TargetUnknown,
			Reasons: []string{ReasonTargetSelectionRequired},
		}, fmt.Errorf("%w", ErrPreflightFailed)
	}

	minSize := req.MinSizeBytes
	if minSize == 0 {
		minSize = MinInstallSizeBytes
	}

	inventory := c.Collector.Collect(context.Background())
	if inventory.Status == storageinventory.StatusUnavailable {
		return TargetResult{
			DiskID:  diskID,
			Status:  TargetUnknown,
			Reasons: []string{ReasonDiscoveryUnavailable},
		}, fmt.Errorf("%w", ErrPreflightFailed)
	}

	device, matches := findDevice(inventory.Devices, diskID)
	switch matches {
	case 0:
		return TargetResult{
			DiskID:  diskID,
			Status:  TargetNotFound,
			Reasons: []string{ReasonTargetNotFound},
		}, fmt.Errorf("%w", ErrPreflightFailed)
	case 1:
	default:
		return TargetResult{
			DiskID:  diskID,
			Status:  TargetAmbiguous,
			Reasons: []string{ReasonTargetAmbiguous},
		}, fmt.Errorf("%w", ErrPreflightFailed)
	}

	result := TargetResult{
		DiskID:    device.ID,
		SizeBytes: device.SizeBytes,
	}

	switch device.Eligibility {
	case storageinventory.EligibilityEligible:
		result.Status = TargetEligible
	case storageinventory.EligibilityExcluded:
		result.Status = TargetExcluded
		result.Reasons = append([]string(nil), device.Reasons...)
	case storageinventory.EligibilityUnknown:
		result.Status = TargetUnknown
		result.Reasons = append([]string(nil), device.Reasons...)
	default:
		result.Status = TargetUnknown
		result.Reasons = []string{storageinventory.ReasonAmbiguousIdentity}
	}

	if device.SizeBytes != nil && *device.SizeBytes < minSize {
		result.Status = TargetExcluded
		result.Reasons = appendUnique(result.Reasons, ReasonInsufficientSize)
	}

	if result.Status != TargetEligible {
		return result, fmt.Errorf("%w", ErrPreflightFailed)
	}
	return result, nil
}

func findDevice(devices []storageinventory.Device, diskID string) (storageinventory.Device, int) {
	var found storageinventory.Device
	count := 0
	for _, device := range devices {
		if device.ID != diskID {
			continue
		}
		found = device
		count++
	}
	return found, count
}

func appendUnique(reasons []string, reason string) []string {
	for _, existing := range reasons {
		if existing == reason {
			return reasons
		}
	}
	return append(reasons, reason)
}
