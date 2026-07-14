package overview_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
	"github.com/crankurbex2025-source/vyntrio-os/internal/application/overview"
	"github.com/crankurbex2025-source/vyntrio-os/internal/application/settings"
	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/setting"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backupstatus"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/hostmetrics"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/netpresence"
)

type mockRepository struct {
	byKey map[string]setting.Setting
}

func (m *mockRepository) Get(_ context.Context, namespace, key string) (setting.Setting, error) {
	s, ok := m.byKey[key]
	if !ok {
		return setting.Setting{}, settings.ErrNotFound
	}
	if s.Namespace != namespace {
		return setting.Setting{}, settings.ErrNotFound
	}
	return s, nil
}

func (m *mockRepository) Set(context.Context, setting.Setting) error {
	return errors.New("not implemented")
}

func (m *mockRepository) ListByNamespace(context.Context, string) ([]setting.Setting, error) {
	return nil, errors.New("not implemented")
}

type stubHostCollector struct{}

func (stubHostCollector) Collect(context.Context) hostmetrics.Host {
	load := 0.12
	cores := 2
	total := uint64(2048 * 1024)
	available := uint64(1024 * 1024)
	used := uint64(1024 * 1024)
	fsType := "ext4"
	fsTotal := uint64(1000)
	fsAvailable := uint64(400)
	fsUsed := uint64(600)
	return hostmetrics.Host{
		CPU: hostmetrics.CPU{
			Status:       hostmetrics.StatusOK,
			LogicalCores: &cores,
			Load1m:       &load,
		},
		Memory: hostmetrics.Memory{
			Status:         hostmetrics.StatusOK,
			TotalBytes:     &total,
			AvailableBytes: &available,
			UsedBytes:      &used,
		},
		Filesystems: []hostmetrics.Filesystem{{
			ID:             hostmetrics.StateFilesystemID,
			Status:         hostmetrics.StatusOK,
			TotalBytes:     &fsTotal,
			AvailableBytes: &fsAvailable,
			UsedBytes:      &fsUsed,
			FSType:         &fsType,
		}},
	}
}

type stubBackupLoader struct {
	status backupstatus.Backup
}

func (s stubBackupLoader) Read(context.Context) backupstatus.Backup {
	return s.status
}

type stubNetworkCollector struct {
	network netpresence.Network
}

func (s stubNetworkCollector) Collect(context.Context) netpresence.Network {
	return s.network
}

type stubReadiness struct {
	result health.Result
}

func (s stubReadiness) Check(context.Context) health.Result {
	return s.result
}

func TestMapReadinessReady(t *testing.T) {
	got := overview.MapReadiness(health.Result{ProcessOK: true, DatabaseOK: true})
	if got.Status != "ready" || got.Database != "ok" {
		t.Fatalf("MapReadiness() = %+v, want ready/ok", got)
	}
}

func TestMapReadinessNotReady(t *testing.T) {
	got := overview.MapReadiness(health.Result{ProcessOK: true, DatabaseOK: false})
	if got.Status != "not_ready" || got.Database != "error" {
		t.Fatalf("MapReadiness() = %+v, want not_ready/error", got)
	}
}

func TestLoaderAssemblesDeterministicOverview(t *testing.T) {
	repo := &mockRepository{
		byKey: map[string]setting.Setting{
			setting.KeyHostname: {
				Namespace: setting.NamespaceSystem,
				Key:       setting.KeyHostname,
				Value:     "Vyntrio Home",
				ValueType: setting.ValueTypeString,
			},
		},
	}
	loader := overview.NewLoader(
		repo,
		stubReadiness{result: health.Result{ProcessOK: true, DatabaseOK: true}},
		stubHostCollector{},
		stubBackupLoader{status: backupstatus.NeverRun()},
		stubNetworkCollector{network: netpresence.Network{Status: netpresence.StatusUnknown}},
		"0.2.0-dev",
		"abc123",
		"development",
	)
	got, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got.Instance.Name != "Vyntrio Home" {
		t.Fatalf("instance.name = %q", got.Instance.Name)
	}
	if got.Instance.Version != "0.2.0-dev" {
		t.Fatalf("instance.version = %q", got.Instance.Version)
	}
	if got.Instance.Commit != "abc123" {
		t.Fatalf("instance.commit = %q", got.Instance.Commit)
	}
	if got.API.Environment != "development" {
		t.Fatalf("api.environment = %q", got.API.Environment)
	}
	if got.Service.Status != "running" {
		t.Fatalf("service.status = %q", got.Service.Status)
	}
	if got.Readiness.Status != "ready" || got.Readiness.Database != "ok" {
		t.Fatalf("readiness = %+v", got.Readiness)
	}
	if got.Backup.Status != backupstatus.StatusNeverRun {
		t.Fatalf("backup = %+v", got.Backup)
	}
	if got.Network.Status != netpresence.StatusUnknown {
		t.Fatalf("network = %+v", got.Network)
	}
	if got.CollectedAt == "" {
		t.Fatal("expected collected_at")
	}
	if _, err := time.Parse(time.RFC3339Nano, got.CollectedAt); err != nil {
		t.Fatalf("collected_at parse error: %v", err)
	}
}

func TestLoaderMapsDatabaseFailureToNotReady(t *testing.T) {
	repo := &mockRepository{
		byKey: map[string]setting.Setting{
			setting.KeyHostname: {
				Namespace: setting.NamespaceSystem,
				Key:       setting.KeyHostname,
				Value:     "Vyntrio Home",
				ValueType: setting.ValueTypeString,
			},
		},
	}
	loader := overview.NewLoader(
		repo,
		stubReadiness{result: health.Result{ProcessOK: true, DatabaseOK: false}},
		stubHostCollector{},
		stubBackupLoader{status: backupstatus.NeverRun()},
		stubNetworkCollector{network: netpresence.Unavailable()},
		"0.2.0-dev",
		"abc123",
		"development",
	)

	got, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got.Readiness.Status != "not_ready" {
		t.Fatalf("readiness.status = %q, want not_ready", got.Readiness.Status)
	}
	if got.Readiness.Database != "error" {
		t.Fatalf("readiness.database = %q, want error", got.Readiness.Database)
	}
	if got.Service.Status != "running" {
		t.Fatalf("service.status = %q, want running", got.Service.Status)
	}
	if got.Network.Status != netpresence.StatusUnavailable {
		t.Fatalf("network.status = %q, want unavailable", got.Network.Status)
	}
}

func TestLoaderAssemblesNetworkAvailable(t *testing.T) {
	repo := &mockRepository{
		byKey: map[string]setting.Setting{
			setting.KeyHostname: {
				Namespace: setting.NamespaceSystem,
				Key:       setting.KeyHostname,
				Value:     "Vyntrio Home",
				ValueType: setting.ValueTypeString,
			},
		},
	}
	loader := overview.NewLoader(
		repo,
		stubReadiness{result: health.Result{ProcessOK: true, DatabaseOK: true}},
		stubHostCollector{},
		stubBackupLoader{status: backupstatus.NeverRun()},
		stubNetworkCollector{network: netpresence.Network{Status: netpresence.StatusAvailable}},
		"0.2.0-dev",
		"abc123",
		"development",
	)

	got, err := loader.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got.Network.Status != netpresence.StatusAvailable {
		t.Fatalf("network.status = %q, want available", got.Network.Status)
	}
}
