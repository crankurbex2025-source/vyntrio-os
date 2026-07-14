//go:build linux

package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backupstatus"
)

const integrationBackupGID uint32 = 4242

func integrationBackupNow() time.Time {
	return time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
}

func writeIntegrationBackupSidecar(t *testing.T, stateDir string, record backupstatus.DiskRecord) {
	t.Helper()

	data, err := backupstatus.EncodeDiskRecord(record)
	if err != nil {
		t.Fatalf("EncodeDiskRecord() error: %v", err)
	}
	path := backupstatus.StatusPath(stateDir)
	if err := os.WriteFile(path, data, 0o640); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
	if err := os.Chown(path, 0, int(integrationBackupGID)); err != nil {
		t.Fatalf("Chown() error: %v", err)
	}
}

func TestOverviewBackupStatusFromRealReader(t *testing.T) {
	stateDir := t.TempDir()
	completedAt := "2026-07-14T11:30:00.000000000Z"
	writeIntegrationBackupSidecar(t, stateDir, backupstatus.DiskRecord{
		SchemaVersion: backupstatus.SchemaVersion,
		LastOutcome:   backupstatus.OutcomeSucceeded,
		CompletedAt:   completedAt,
		EverSucceeded: true,
	})

	reader := backupstatus.NewReader(stateDir, backupstatus.ReaderDeps{
		GroupGID: func() (uint32, error) { return integrationBackupGID, nil },
		Now:      integrationBackupNow,
	})
	router := newSettingsRouter(t, settingsRouterOpts{backupStatus: reader})
	sessionCookie := ownerSessionCookie(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	got := decodeOverviewResponse(t, rec.Body.Bytes())
	if got.Backup.Status != backupstatus.StatusSucceeded {
		t.Fatalf("backup.status = %q, want succeeded", got.Backup.Status)
	}
	if got.Backup.CompletedAt == nil || *got.Backup.CompletedAt != completedAt {
		t.Fatalf("backup.completed_at = %+v, want %q", got.Backup.CompletedAt, completedAt)
	}
	if got.Backup.EverSucceeded == nil || !*got.Backup.EverSucceeded {
		t.Fatalf("backup.ever_succeeded = %+v, want true", got.Backup.EverSucceeded)
	}
	if got.Backup.Failure != nil {
		t.Fatalf("backup.failure = %+v, want omitted", got.Backup.Failure)
	}

	body := rec.Body.String()
	if strings.Contains(body, stateDir) {
		t.Fatalf("response leaked temp path: %s", body)
	}
	assertOverviewCacheControlNoStore(t, rec)
	assertNoSensitiveOverviewFields(t, rec.Body.Bytes())
}
