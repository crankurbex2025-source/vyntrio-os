package overview

import (
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backupstatus"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/hostmetrics"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/netpresence"
)

// Response is the safe authenticated overview DTO for GET /api/v1/overview.
type Response struct {
	Instance    InstanceSection     `json:"instance"`
	API         APISection          `json:"api"`
	Service     ServiceSection      `json:"service"`
	Readiness   ReadinessSection    `json:"readiness"`
	Host        hostmetrics.Host    `json:"host"`
	Backup      backupstatus.Backup `json:"backup"`
	Network     netpresence.Network `json:"network"`
	CollectedAt string              `json:"collected_at"`
}

// InstanceSection exposes non-secret appliance identity metadata.
type InstanceSection struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// APISection exposes non-secret API runtime metadata.
type APISection struct {
	Environment string `json:"environment"`
}

// ServiceSection reports in-process service liveness for the overview handler.
type ServiceSection struct {
	Status string `json:"status"`
}

// ReadinessSection reports database readiness using the same semantics as /readyz.
type ReadinessSection struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}
