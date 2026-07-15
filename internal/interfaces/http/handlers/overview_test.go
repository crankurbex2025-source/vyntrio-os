package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appoverview "github.com/crankurbex2025-source/vyntrio-os/internal/application/overview"
	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/backupstatus"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/hostmetrics"
	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/netpresence"
)

const (
	overviewPath            = "/api/v1/overview"
	overviewTestVersion     = "0.2.0-test"
	overviewTestEnvironment = "development"
)

func overviewGET(cookies []*http.Cookie) *http.Request {
	req := httptest.NewRequest(http.MethodGet, overviewPath, nil)
	req.RemoteAddr = "127.0.0.1:8080"
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return req
}

func TestOverviewUnauthenticatedReturns401(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET(nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if parseSettingsErrorCode(t, rec.Body.Bytes()) != "UNAUTHORIZED" {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestOverviewMissingPermissionReturns403(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{
		authorizer: denyAuthorizer{},
	})

	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if parseSettingsErrorCode(t, rec.Body.Bytes()) != "FORBIDDEN" {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestOverviewRealRBACAuthorizationMatrix(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})

	cases := []struct {
		name       string
		userID     domainidentity.UserID
		username   string
		role       domainidentity.Role
		sessionID  string
		rawToken   string
		wantStatus int
	}{
		{
			name:       "owner",
			userID:     domainidentity.UserID("overview-owner"),
			username:   "overview-owner",
			role:       domainidentity.RoleOwner,
			sessionID:  "overview-owner-sess",
			rawToken:   "overview-owner-token",
			wantStatus: http.StatusOK,
		},
		{
			name:       "operator",
			userID:     domainidentity.UserID("overview-operator"),
			username:   "overview-operator",
			role:       domainidentity.RoleOperator,
			sessionID:  "overview-operator-sess",
			rawToken:   "overview-operator-token",
			wantStatus: http.StatusOK,
		},
		{
			name:       "read_only",
			userID:     domainidentity.UserID("overview-read-only"),
			username:   "overview-read-only",
			role:       domainidentity.RoleReadOnly,
			sessionID:  "overview-read-only-sess",
			rawToken:   "overview-read-only-token",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sessionCookie := createRoleSettingsSession(
				t,
				router,
				tc.userID,
				tc.username,
				tc.role,
				tc.sessionID,
				tc.rawToken,
			)

			rec := httptest.NewRecorder()
			router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d body=%s, want %d", rec.Code, rec.Body.String(), tc.wantStatus)
			}
			if tc.wantStatus != http.StatusOK {
				return
			}

			assertOverviewResponseShape(t, rec)
			assertOverviewCacheControlNoStore(t, rec)
			assertNoSensitiveOverviewFields(t, rec.Body.Bytes())
		})
	}
}

func TestOverviewResponseShapeOwnerSession(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	sessionCookie := ownerSessionCookie(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	got := decodeOverviewResponse(t, rec.Body.Bytes())
	if got.Instance.Name == "" {
		t.Fatal("expected instance name")
	}
	if got.Instance.Version != settingsTestVersion {
		t.Fatalf("version = %q, want %q", got.Instance.Version, settingsTestVersion)
	}
	if got.Instance.Commit != "test-commit" {
		t.Fatalf("commit = %q, want test-commit", got.Instance.Commit)
	}
	if got.API.Environment != settingsTestEnvironment {
		t.Fatalf("environment = %q, want %q", got.API.Environment, settingsTestEnvironment)
	}
	if got.Service.Status != "running" {
		t.Fatalf("service.status = %q, want running", got.Service.Status)
	}
	if got.Readiness.Status != "ready" {
		t.Fatalf("readiness.status = %q, want ready", got.Readiness.Status)
	}
	if got.Readiness.Database != "ok" {
		t.Fatalf("readiness.database = %q, want ok", got.Readiness.Database)
	}
	if got.CollectedAt == "" {
		t.Fatal("expected collected_at")
	}
	if _, err := time.Parse(time.RFC3339Nano, got.CollectedAt); err != nil {
		t.Fatalf("collected_at parse error: %v", err)
	}

	assertOverviewCacheControlNoStore(t, rec)
	assertNoSensitiveOverviewFields(t, rec.Body.Bytes())
}

func TestOverviewDatabaseFailureReturns200NotReady(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{
		readinessDB: failingDBChecker{},
	})

	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s, want 200", rec.Code, rec.Body.String())
	}

	got := decodeOverviewResponse(t, rec.Body.Bytes())
	if got.Readiness.Status != "not_ready" {
		t.Fatalf("readiness.status = %q, want not_ready", got.Readiness.Status)
	}
	if got.Readiness.Database != "error" {
		t.Fatalf("readiness.database = %q, want error", got.Readiness.Database)
	}
	if got.Service.Status != "running" {
		t.Fatalf("service.status = %q, want running", got.Service.Status)
	}
}

func TestOverviewDegradedHostMetricsReturn200WithUnavailableSections(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{
		hostMetrics: unavailableHostMetricsCollector{},
	})
	sessionCookie := ownerSessionCookie(t, router)

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s, want 200", rec.Code, rec.Body.String())
	}

	assertOverviewCacheControlNoStore(t, rec)
	assertNoSensitiveOverviewFields(t, rec.Body.Bytes())
	assertDegradedHostMetricsJSON(t, rec.Body.Bytes())

	got := decodeOverviewResponse(t, rec.Body.Bytes())
	if got.Readiness.Status != "ready" || got.Readiness.Database != "ok" {
		t.Fatalf("readiness = %+v, want ready/ok", got.Readiness)
	}
	if got.Service.Status != "running" {
		t.Fatalf("service.status = %q, want running", got.Service.Status)
	}
}

type unavailableHostMetricsCollector struct{}

func (unavailableHostMetricsCollector) Collect(context.Context) hostmetrics.Host {
	return hostmetrics.Host{
		CPU:    hostmetrics.CPU{Status: hostmetrics.StatusUnavailable},
		Memory: hostmetrics.Memory{Status: hostmetrics.StatusUnavailable},
		Filesystems: []hostmetrics.Filesystem{{
			ID:     hostmetrics.StateFilesystemID,
			Status: hostmetrics.StatusUnavailable,
		}},
	}
}

func assertDegradedHostMetricsJSON(t *testing.T, body []byte) {
	t.Helper()

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error: %v body=%s", err, body)
	}
	hostRaw, ok := payload["host"]
	if !ok {
		t.Fatalf("missing host in %s", body)
	}

	var host map[string]json.RawMessage
	if err := json.Unmarshal(hostRaw, &host); err != nil {
		t.Fatalf("json.Unmarshal(host) error: %v", err)
	}

	assertUnavailableMetricSection(t, "cpu", host["cpu"], nil)
	assertUnavailableMetricSection(t, "memory", host["memory"], nil)

	var filesystems []json.RawMessage
	if err := json.Unmarshal(host["filesystems"], &filesystems); err != nil {
		t.Fatalf("json.Unmarshal(filesystems) error: %v", err)
	}
	if len(filesystems) != 1 {
		t.Fatalf("filesystems length = %d, want 1", len(filesystems))
	}
	assertUnavailableMetricSection(t, "filesystems[0]", filesystems[0], map[string]struct{}{"id": {}})

	bodyLower := strings.ToLower(string(body))
	for _, forbidden := range []string{
		`"logical_cores":0`, `"logical_cores":null`,
		`"load_1m":0`, `"load_1m":null`,
		`"total_bytes":0`, `"total_bytes":null`,
		`"available_bytes":0`, `"available_bytes":null`,
		`"used_bytes":0`, `"used_bytes":null`,
	} {
		if strings.Contains(bodyLower, forbidden) {
			t.Fatalf("response contained forbidden numeric host field %q: %s", forbidden, body)
		}
	}
}

func assertUnavailableMetricSection(
	t *testing.T,
	label string,
	raw json.RawMessage,
	allowedExtraKeys map[string]struct{},
) {
	t.Helper()

	var section map[string]json.RawMessage
	if err := json.Unmarshal(raw, &section); err != nil {
		t.Fatalf("json.Unmarshal(%s) error: %v", label, err)
	}
	if len(section) != 1+len(allowedExtraKeys) {
		t.Fatalf("%s keys = %v, want status-only section", label, sectionKeys(section))
	}
	if string(section["status"]) != `"unavailable"` {
		t.Fatalf("%s status = %s, want unavailable", label, section["status"])
	}
	for key := range section {
		if key == "status" {
			continue
		}
		if _, ok := allowedExtraKeys[key]; !ok {
			t.Fatalf("%s contained unexpected key %q", label, key)
		}
	}
	for key := range allowedExtraKeys {
		if _, ok := section[key]; !ok {
			t.Fatalf("%s missing required key %q", label, key)
		}
	}
}

func sectionKeys(section map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(section))
	for key := range section {
		keys = append(keys, key)
	}
	return keys
}

func TestOverviewBackupStatusNeverRun(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	got := decodeOverviewResponse(t, rec.Body.Bytes())
	if got.Backup.Status != backupstatus.StatusNeverRun || got.Backup.EverSucceeded == nil || *got.Backup.EverSucceeded {
		t.Fatalf("backup = %+v", got.Backup)
	}
	assertOverviewCacheControlNoStore(t, rec)
	assertNoSensitiveOverviewFields(t, rec.Body.Bytes())
}

func TestOverviewBackupStatusSucceededSerialization(t *testing.T) {
	completedAt := "2026-07-14T11:30:00.000000000Z"
	ever := true
	router := newSettingsRouter(t, settingsRouterOpts{
		backupStatus: stubBackupStatusLoader{status: backupstatus.Backup{
			Status:        backupstatus.StatusSucceeded,
			CompletedAt:   &completedAt,
			EverSucceeded: &ever,
		}},
	})
	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	assertBackupJSON(t, rec.Body.Bytes(), `"status":"succeeded"`, `"ever_succeeded":true`)
	assertOverviewCacheControlNoStore(t, rec)
	assertNoSensitiveOverviewFields(t, rec.Body.Bytes())
}

func TestOverviewBackupStatusFailedFirstRunSerialization(t *testing.T) {
	completedAt := "2026-07-14T11:30:00.000000000Z"
	ever := false
	failure := backupstatus.FailureArtifact
	router := newSettingsRouter(t, settingsRouterOpts{
		backupStatus: stubBackupStatusLoader{status: backupstatus.Backup{
			Status:        backupstatus.StatusFailed,
			CompletedAt:   &completedAt,
			EverSucceeded: &ever,
			Failure:       &failure,
		}},
	})
	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	assertOverviewCacheControlNoStore(t, rec)
	assertBackupFailedJSON(t, rec.Body.Bytes(), completedAt, false, backupstatus.FailureArtifact)
	assertNoSensitiveOverviewFields(t, rec.Body.Bytes())
	assertOverviewResponseShape(t, rec)
}

func TestOverviewBackupStatusFailedAfterPriorSuccessSerialization(t *testing.T) {
	completedAt := "2026-07-14T12:45:00.000000000Z"
	ever := true
	failure := backupstatus.FailureReadiness
	router := newSettingsRouter(t, settingsRouterOpts{
		backupStatus: stubBackupStatusLoader{status: backupstatus.Backup{
			Status:        backupstatus.StatusFailed,
			CompletedAt:   &completedAt,
			EverSucceeded: &ever,
			Failure:       &failure,
		}},
	})
	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	assertOverviewCacheControlNoStore(t, rec)
	assertBackupFailedJSON(t, rec.Body.Bytes(), completedAt, true, backupstatus.FailureReadiness)
	assertNoSensitiveOverviewFields(t, rec.Body.Bytes())
	assertOverviewResponseShape(t, rec)
}

func TestOverviewBackupStatusUnavailableOmitsFields(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{
		backupStatus: stubBackupStatusLoader{status: backupstatus.Unavailable()},
	})
	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	got := decodeOverviewResponse(t, rec.Body.Bytes())
	if got.Backup.Status != backupstatus.StatusUnavailable || got.Backup.EverSucceeded != nil {
		t.Fatalf("backup = %+v", got.Backup)
	}
	bodyLower := strings.ToLower(rec.Body.String())
	for _, forbidden := range []string{`"completed_at"`, `"failure"`, `"ever_succeeded"`} {
		if strings.Contains(bodyLower, forbidden) {
			t.Fatalf("response contained %q: %s", forbidden, rec.Body.String())
		}
	}
}

type stubBackupStatusLoader struct {
	status backupstatus.Backup
}

func (s stubBackupStatusLoader) Read(context.Context) backupstatus.Backup {
	return s.status
}

type stubNetworkPresenceLoader struct {
	network netpresence.Network
}

func (s stubNetworkPresenceLoader) Collect(context.Context) netpresence.Network {
	return s.network
}

func assertBackupJSON(t *testing.T, body []byte, wantContains ...string) {
	t.Helper()
	lower := strings.ToLower(string(body))
	for _, want := range wantContains {
		if !strings.Contains(lower, strings.ToLower(want)) {
			t.Fatalf("body missing %q: %s", want, body)
		}
	}
}

func assertBackupFailedJSON(
	t *testing.T,
	body []byte,
	wantCompletedAt string,
	wantEverSucceeded bool,
	wantFailure string,
) {
	t.Helper()

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error: %v body=%s", err, body)
	}
	backupRaw, ok := payload["backup"]
	if !ok {
		t.Fatalf("missing backup in %s", body)
	}

	var backup map[string]json.RawMessage
	if err := json.Unmarshal(backupRaw, &backup); err != nil {
		t.Fatalf("json.Unmarshal(backup) error: %v", err)
	}

	wantKeys := map[string]struct{}{
		"status": {}, "completed_at": {}, "ever_succeeded": {}, "failure": {},
	}
	if len(backup) != len(wantKeys) {
		t.Fatalf("backup keys = %v, want exactly %v", sectionKeys(backup), wantKeys)
	}
	for key := range wantKeys {
		if _, ok := backup[key]; !ok {
			t.Fatalf("backup missing key %q", key)
		}
	}

	if string(backup["status"]) != `"failed"` {
		t.Fatalf("backup.status = %s, want failed", backup["status"])
	}
	if string(backup["completed_at"]) != `"`+wantCompletedAt+`"` {
		t.Fatalf("backup.completed_at = %s, want %q", backup["completed_at"], wantCompletedAt)
	}
	wantEver := "false"
	if wantEverSucceeded {
		wantEver = "true"
	}
	if string(backup["ever_succeeded"]) != wantEver {
		t.Fatalf("backup.ever_succeeded = %s, want %s", backup["ever_succeeded"], wantEver)
	}
	if string(backup["failure"]) != `"`+wantFailure+`"` {
		t.Fatalf("backup.failure = %s, want %q", backup["failure"], wantFailure)
	}
}

func TestOverviewPreservesHealthEndpointBehavior(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})

	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("readyz status = %d, want 200", rec.Code)
	}
}

type failingDBChecker struct{}

func (failingDBChecker) Ping(context.Context) error {
	return context.Canceled
}

func decodeOverviewResponse(t *testing.T, body []byte) appoverview.Response {
	t.Helper()
	var got appoverview.Response
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("json.Unmarshal() error: %v body=%s", err, body)
	}
	return got
}

func assertOverviewResponseShape(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	got := decodeOverviewResponse(t, rec.Body.Bytes())
	if got.Instance.Name == "" || got.Instance.Version == "" || got.Instance.Commit == "" {
		t.Fatalf("instance = %+v", got.Instance)
	}
	if got.API.Environment == "" {
		t.Fatalf("api = %+v", got.API)
	}
	if got.Service.Status != "running" {
		t.Fatalf("service = %+v", got.Service)
	}
	if got.Readiness.Status == "" || got.Readiness.Database == "" {
		t.Fatalf("readiness = %+v", got.Readiness)
	}
	if got.CollectedAt == "" {
		t.Fatal("missing collected_at")
	}
	if got.Host.CPU.Status == "" {
		t.Fatalf("host.cpu = %+v", got.Host.CPU)
	}
	if got.Host.Memory.Status == "" {
		t.Fatalf("host.memory = %+v", got.Host.Memory)
	}
	if len(got.Host.Filesystems) != 1 || got.Host.Filesystems[0].ID != "state" {
		t.Fatalf("host.filesystems = %+v", got.Host.Filesystems)
	}
	if got.Backup.Status == "" {
		t.Fatalf("backup = %+v", got.Backup)
	}
	if got.Network.Status == "" {
		t.Fatalf("network = %+v", got.Network)
	}
	if got.Software.Status == "" {
		t.Fatalf("software = %+v", got.Software)
	}
	if got.Runtime.Status == "" {
		t.Fatalf("runtime = %+v", got.Runtime)
	}
	if got.Health.Status == "" {
		t.Fatalf("health = %+v", got.Health)
	}
}

func assertOverviewCacheControlNoStore(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}

var forbiddenOverviewJSONKeys = map[string]struct{}{
	"password": {}, "token": {}, "hash": {}, "csrf": {}, "session": {},
	"userid": {}, "user_id": {}, "role": {}, "principal": {},
	"datadir": {}, "data_dir": {}, "path": {}, "port": {}, "bind": {},
	"timezone": {}, "audit": {}, "cookie": {}, "credential": {},
	"destination": {}, "preserve": {}, "digest": {}, "manifest": {},
	"stderr": {}, "stdout": {}, "source_path": {}, "artifact_name": {},
	"config_path": {}, "state_root": {}, "config.toml": {},
	"interface": {}, "mac": {}, "address": {}, "ip": {},
	"route": {}, "gateway": {}, "dns": {}, "mtu": {}, "index": {},
	"count": {}, "error": {},
}

var allowedBackupFailureClasses = map[string]struct{}{
	backupstatus.FailureArtifact:  {},
	backupstatus.FailureRestart:   {},
	backupstatus.FailureHealth:    {},
	backupstatus.FailureReadiness: {},
	backupstatus.FailureInternal:  {},
}

var allowedNetworkStatuses = map[string]struct{}{
	netpresence.StatusAvailable:   {},
	netpresence.StatusUnknown:     {},
	netpresence.StatusUnavailable: {},
}

var allowedSoftwareStatuses = map[string]struct{}{
	appoverview.SoftwareStatusOK:          {},
	appoverview.SoftwareStatusUnavailable: {},
}

var allowedSoftwareChannels = map[string]struct{}{
	appoverview.ReleaseChannelDevelopment: {},
	appoverview.ReleaseChannelProduction:  {},
	appoverview.ReleaseChannelUnknown:     {},
}

var allowedRuntimeStatuses = map[string]struct{}{
	appoverview.RuntimeStatusReady:    {},
	appoverview.RuntimeStatusDegraded: {},
	appoverview.RuntimeStatusUnknown:  {},
}

var allowedRuntimeNotes = map[string]struct{}{
	appoverview.RuntimeNoteDatabase: {},
}

var allowedHealthStatuses = map[string]struct{}{
	appoverview.HealthStatusHealthy: {},
	appoverview.HealthStatusWarning: {},
	appoverview.HealthStatusUnknown: {},
}

var allowedHealthNotes = map[string]struct{}{
	appoverview.HealthNoteDatabase: {},
	appoverview.HealthNoteBackup:   {},
}

var forbiddenOverviewValueSubstrings = []string{
	"/var/lib", "/etc/vyntrio", "127.0.0.1", "sqlite", "config.toml",
	".tar", "sha256", "vyntrio.db",
}

func assertNoSensitiveOverviewFields(t *testing.T, body []byte) {
	t.Helper()
	if err := checkNoSensitiveOverviewFields(body); err != nil {
		t.Fatal(err)
	}
}

func checkNoSensitiveOverviewFields(body []byte) error {
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	return scanOverviewJSONValue(payload, nil)
}

func scanOverviewJSONValue(value any, parentKeys []string) error {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			lowerKey := strings.ToLower(key)
			if lowerKey == "name" && !isInstanceSection(parentKeys) {
				return fmt.Errorf("response contained forbidden JSON key %q", key)
			}
			if _, forbidden := forbiddenOverviewJSONKeys[lowerKey]; forbidden {
				return fmt.Errorf("response contained forbidden JSON key %q", key)
			}
			if isNetworkSection(parentKeys) && lowerKey != "status" {
				return fmt.Errorf("network section contained forbidden JSON key %q", key)
			}
			if isSoftwareSection(parentKeys) {
				allowed := map[string]struct{}{
					"status": {}, "version": {}, "commit": {}, "channel": {},
				}
				if _, ok := allowed[lowerKey]; !ok {
					return fmt.Errorf("software section contained forbidden JSON key %q", key)
				}
			}
			if isRuntimeSection(parentKeys) {
				allowed := map[string]struct{}{
					"status": {}, "note": {},
				}
				if _, ok := allowed[lowerKey]; !ok {
					return fmt.Errorf("runtime section contained forbidden JSON key %q", key)
				}
			}
			if isHealthSection(parentKeys) {
				allowed := map[string]struct{}{
					"status": {}, "note": {},
				}
				if _, ok := allowed[lowerKey]; !ok {
					return fmt.Errorf("health section contained forbidden JSON key %q", key)
				}
			}
			if lowerKey == "failure" && isBackupSection(parentKeys) {
				failure, ok := child.(string)
				if !ok {
					return fmt.Errorf("backup.failure has unexpected type %T", child)
				}
				if _, allowed := allowedBackupFailureClasses[failure]; !allowed {
					return fmt.Errorf("backup.failure = %q, want allowed enum", failure)
				}
				continue
			}
			if lowerKey == "status" && isNetworkSection(parentKeys) {
				status, ok := child.(string)
				if !ok {
					return fmt.Errorf("network.status has unexpected type %T", child)
				}
				if _, allowed := allowedNetworkStatuses[status]; !allowed {
					return fmt.Errorf("network.status = %q, want allowed enum", status)
				}
				continue
			}
			if lowerKey == "status" && isSoftwareSection(parentKeys) {
				status, ok := child.(string)
				if !ok {
					return fmt.Errorf("software.status has unexpected type %T", child)
				}
				if _, allowed := allowedSoftwareStatuses[status]; !allowed {
					return fmt.Errorf("software.status = %q, want allowed enum", status)
				}
				continue
			}
			if lowerKey == "channel" && isSoftwareSection(parentKeys) {
				channel, ok := child.(string)
				if !ok {
					return fmt.Errorf("software.channel has unexpected type %T", child)
				}
				if _, allowed := allowedSoftwareChannels[channel]; !allowed {
					return fmt.Errorf("software.channel = %q, want allowed enum", channel)
				}
				continue
			}
			if lowerKey == "status" && isRuntimeSection(parentKeys) {
				status, ok := child.(string)
				if !ok {
					return fmt.Errorf("runtime.status has unexpected type %T", child)
				}
				if _, allowed := allowedRuntimeStatuses[status]; !allowed {
					return fmt.Errorf("runtime.status = %q, want allowed enum", status)
				}
				continue
			}
			if lowerKey == "note" && isRuntimeSection(parentKeys) {
				note, ok := child.(string)
				if !ok {
					return fmt.Errorf("runtime.note has unexpected type %T", child)
				}
				if _, allowed := allowedRuntimeNotes[note]; !allowed {
					return fmt.Errorf("runtime.note = %q, want allowed enum", note)
				}
				continue
			}
			if lowerKey == "status" && isHealthSection(parentKeys) {
				status, ok := child.(string)
				if !ok {
					return fmt.Errorf("health.status has unexpected type %T", child)
				}
				if _, allowed := allowedHealthStatuses[status]; !allowed {
					return fmt.Errorf("health.status = %q, want allowed enum", status)
				}
				continue
			}
			if lowerKey == "note" && isHealthSection(parentKeys) {
				note, ok := child.(string)
				if !ok {
					return fmt.Errorf("health.note has unexpected type %T", child)
				}
				if _, allowed := allowedHealthNotes[note]; !allowed {
					return fmt.Errorf("health.note = %q, want allowed enum", note)
				}
				continue
			}
			if err := scanOverviewJSONValue(child, append(parentKeys, lowerKey)); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range typed {
			if err := scanOverviewJSONValue(child, parentKeys); err != nil {
				return err
			}
		}
	case string:
		if err := checkForbiddenOverviewStringValue(typed); err != nil {
			return err
		}
	}
	return nil
}

func checkForbiddenOverviewStringValue(value string) error {
	lower := strings.ToLower(value)
	for _, forbidden := range forbiddenOverviewValueSubstrings {
		if strings.Contains(lower, forbidden) {
			return fmt.Errorf("response leaked forbidden value substring %q in %q", forbidden, value)
		}
	}
	if looksLikeIPv4(value) || looksLikeMAC(value) {
		return fmt.Errorf("response leaked address-like value %q", value)
	}
	return nil
}

func looksLikeIPv4(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				return false
			}
		}
	}
	return true
}

func looksLikeMAC(value string) bool {
	parts := strings.Split(value, ":")
	if len(parts) != 6 {
		return false
	}
	for _, part := range parts {
		if len(part) != 2 {
			return false
		}
		for _, ch := range part {
			if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') && (ch < 'A' || ch > 'F') {
				return false
			}
		}
	}
	return true
}

func isBackupSection(parentKeys []string) bool {
	return len(parentKeys) == 1 && parentKeys[0] == "backup"
}

func isInstanceSection(parentKeys []string) bool {
	return len(parentKeys) == 1 && parentKeys[0] == "instance"
}

func isNetworkSection(parentKeys []string) bool {
	return len(parentKeys) == 1 && parentKeys[0] == "network"
}

func isSoftwareSection(parentKeys []string) bool {
	return len(parentKeys) == 1 && parentKeys[0] == "software"
}

func isRuntimeSection(parentKeys []string) bool {
	return len(parentKeys) == 1 && parentKeys[0] == "runtime"
}

func isHealthSection(parentKeys []string) bool {
	return len(parentKeys) == 1 && parentKeys[0] == "health"
}

func assertNetworkJSON(t *testing.T, body []byte, wantStatus string) {
	t.Helper()

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error: %v body=%s", err, body)
	}
	networkRaw, ok := payload["network"]
	if !ok {
		t.Fatalf("missing network in %s", body)
	}
	var network map[string]json.RawMessage
	if err := json.Unmarshal(networkRaw, &network); err != nil {
		t.Fatalf("json.Unmarshal(network) error: %v", err)
	}
	if len(network) != 1 {
		t.Fatalf("network keys = %v, want status only", sectionKeys(network))
	}
	if string(network["status"]) != `"`+wantStatus+`"` {
		t.Fatalf("network.status = %s, want %q", network["status"], wantStatus)
	}
}

func TestOverviewNetworkStatusAvailableSerialization(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{
		networkPresence: stubNetworkPresenceLoader{network: netpresence.Network{Status: netpresence.StatusAvailable}},
	})
	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	assertOverviewCacheControlNoStore(t, rec)
	assertNetworkJSON(t, rec.Body.Bytes(), netpresence.StatusAvailable)
	assertNoSensitiveOverviewFields(t, rec.Body.Bytes())
}

func TestOverviewNetworkStatusUnknownSerialization(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{
		networkPresence: stubNetworkPresenceLoader{network: netpresence.Network{Status: netpresence.StatusUnknown}},
	})
	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	assertNetworkJSON(t, rec.Body.Bytes(), netpresence.StatusUnknown)
}

func TestOverviewNetworkStatusUnavailableWithReadyReadiness(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{
		networkPresence: stubNetworkPresenceLoader{network: netpresence.Unavailable()},
	})
	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	got := decodeOverviewResponse(t, rec.Body.Bytes())
	if got.Readiness.Status != "ready" || got.Readiness.Database != "ok" {
		t.Fatalf("readiness = %+v", got.Readiness)
	}
	assertNetworkJSON(t, rec.Body.Bytes(), netpresence.StatusUnavailable)
}

func TestOverviewSoftwareStatusOKSerialization(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	assertSoftwareJSON(t, rec.Body.Bytes(), appoverview.SoftwareStatusOK, settingsTestVersion, "test-commit", appoverview.ReleaseChannelDevelopment)
	assertNoSensitiveOverviewFields(t, rec.Body.Bytes())
}

func TestOverviewRuntimeStatusDegradedSerialization(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{
		readinessDB: failingDBChecker{},
	})
	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	got := decodeOverviewResponse(t, rec.Body.Bytes())
	if got.Runtime.Status != appoverview.RuntimeStatusDegraded || got.Runtime.Note != appoverview.RuntimeNoteDatabase {
		t.Fatalf("runtime = %+v, want degraded/database", got.Runtime)
	}
	if got.Readiness.Status != "not_ready" || got.Readiness.Database != "error" {
		t.Fatalf("readiness = %+v", got.Readiness)
	}
	if got.Health.Status != appoverview.HealthStatusWarning || got.Health.Note != appoverview.HealthNoteDatabase {
		t.Fatalf("health = %+v, want warning/database", got.Health)
	}
}

func TestOverviewHealthStatusHealthySerialization(t *testing.T) {
	router := newSettingsRouter(t, settingsRouterOpts{})
	sessionCookie := ownerSessionCookie(t, router)
	rec := httptest.NewRecorder()
	router.handler.ServeHTTP(rec, overviewGET([]*http.Cookie{sessionCookie}))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	got := decodeOverviewResponse(t, rec.Body.Bytes())
	if got.Health.Status != appoverview.HealthStatusHealthy {
		t.Fatalf("health = %+v, want healthy", got.Health)
	}
}

func assertSoftwareJSON(t *testing.T, body []byte, wantStatus, wantVersion, wantCommit, wantChannel string) {
	t.Helper()

	got := decodeOverviewResponse(t, body)
	if got.Software.Status != wantStatus {
		t.Fatalf("software.status = %q, want %q", got.Software.Status, wantStatus)
	}
	if wantStatus == appoverview.SoftwareStatusOK {
		if got.Software.Version != wantVersion {
			t.Fatalf("software.version = %q, want %q", got.Software.Version, wantVersion)
		}
		if got.Software.Commit != wantCommit {
			t.Fatalf("software.commit = %q, want %q", got.Software.Commit, wantCommit)
		}
		if got.Software.Channel != wantChannel {
			t.Fatalf("software.channel = %q, want %q", got.Software.Channel, wantChannel)
		}
	}
}

func TestDisclosureGuardAllowsArtifactFailureClass(t *testing.T) {
	body := []byte(`{
		"instance":{"name":"Vyntrio Home","version":"0.2.0-test","commit":"test-commit"},
		"api":{"environment":"development"},
		"service":{"status":"running"},
		"readiness":{"status":"ready","database":"ok"},
		"host":{"cpu":{"status":"unavailable"},"memory":{"status":"unavailable"},"filesystems":[{"id":"state","status":"unavailable"}]},
		"backup":{"status":"failed","completed_at":"2026-07-14T11:30:00.000000000Z","ever_succeeded":false,"failure":"artifact"},
		"network":{"status":"available"},
		"software":{"status":"ok","version":"0.2.0-test","commit":"test-commit","channel":"development"},
		"runtime":{"status":"ready"},
		"health":{"status":"warning","note":"backup"},
		"collected_at":"2026-07-14T12:00:00.000000000Z"
	}`)
	assertNoSensitiveOverviewFields(t, body)
}

func TestDisclosureGuardAllowsNetworkAvailableStatus(t *testing.T) {
	body := []byte(`{
		"instance":{"name":"Vyntrio Home","version":"0.2.0-test","commit":"test-commit"},
		"api":{"environment":"development"},
		"service":{"status":"running"},
		"readiness":{"status":"ready","database":"ok"},
		"host":{"cpu":{"status":"unavailable"},"memory":{"status":"unavailable"},"filesystems":[{"id":"state","status":"unavailable"}]},
		"backup":{"status":"never_run","ever_succeeded":false},
		"network":{"status":"available"},
		"software":{"status":"ok","version":"0.2.0-test","commit":"test-commit","channel":"development"},
		"runtime":{"status":"ready"},
		"health":{"status":"healthy"},
		"collected_at":"2026-07-14T12:00:00.000000000Z"
	}`)
	assertNoSensitiveOverviewFields(t, body)
}

func TestDisclosureGuardRejectsForbiddenInterfaceKey(t *testing.T) {
	body := []byte(`{
		"instance":{"name":"Vyntrio Home","version":"0.2.0-test","commit":"test-commit"},
		"api":{"environment":"development"},
		"service":{"status":"running"},
		"readiness":{"status":"ready","database":"ok"},
		"host":{"cpu":{"status":"unavailable"},"memory":{"status":"unavailable"},"filesystems":[{"id":"state","status":"unavailable"}]},
		"backup":{"status":"never_run","ever_succeeded":false},
		"network":{"status":"available","interface":"eth0"},
		"software":{"status":"ok","version":"0.2.0-test","commit":"test-commit","channel":"development"},
		"runtime":{"status":"ready"},
		"health":{"status":"healthy"},
		"collected_at":"2026-07-14T12:00:00.000000000Z"
	}`)
	if err := checkNoSensitiveOverviewFields(body); err == nil {
		t.Fatal("expected disclosure guard failure for forbidden interface key")
	}
}

func TestDisclosureGuardRejectsForbiddenNameInNetworkSection(t *testing.T) {
	body := []byte(`{
		"instance":{"name":"Vyntrio Home","version":"0.2.0-test","commit":"test-commit"},
		"api":{"environment":"development"},
		"service":{"status":"running"},
		"readiness":{"status":"ready","database":"ok"},
		"host":{"cpu":{"status":"unavailable"},"memory":{"status":"unavailable"},"filesystems":[{"id":"state","status":"unavailable"}]},
		"backup":{"status":"never_run","ever_succeeded":false},
		"network":{"status":"available","name":"eth0"},
		"software":{"status":"ok","version":"0.2.0-test","commit":"test-commit","channel":"development"},
		"runtime":{"status":"ready"},
		"health":{"status":"healthy"},
		"collected_at":"2026-07-14T12:00:00.000000000Z"
	}`)
	if err := checkNoSensitiveOverviewFields(body); err == nil {
		t.Fatal("expected disclosure guard failure for forbidden name key in network section")
	}
}

func TestDisclosureGuardRejectsIPAddressValue(t *testing.T) {
	body := []byte(`{
		"instance":{"name":"Vyntrio Home","version":"0.2.0-test","commit":"test-commit"},
		"api":{"environment":"development"},
		"service":{"status":"running"},
		"readiness":{"status":"ready","database":"ok"},
		"host":{"cpu":{"status":"unavailable"},"memory":{"status":"unavailable"},"filesystems":[{"id":"state","status":"unavailable"}]},
		"backup":{"status":"never_run","ever_succeeded":false},
		"network":{"status":"available","detail":"192.168.1.10"},
		"software":{"status":"ok","version":"0.2.0-test","commit":"test-commit","channel":"development"},
		"runtime":{"status":"ready"},
		"health":{"status":"healthy"},
		"collected_at":"2026-07-14T12:00:00.000000000Z"
	}`)
	if err := checkNoSensitiveOverviewFields(body); err == nil {
		t.Fatal("expected disclosure guard failure for IP-like value")
	}
}

func TestDisclosureGuardRejectsForbiddenPathField(t *testing.T) {
	body := []byte(`{
		"instance":{"name":"Vyntrio Home","version":"0.2.0-test","commit":"test-commit"},
		"api":{"environment":"development"},
		"service":{"status":"running"},
		"readiness":{"status":"ready","database":"ok"},
		"host":{"cpu":{"status":"unavailable"},"memory":{"status":"unavailable"},"filesystems":[{"id":"state","status":"unavailable"}]},
		"backup":{"status":"failed","completed_at":"2026-07-14T11:30:00.000000000Z","ever_succeeded":false,"failure":"artifact","path":"/var/lib/vyntrio/backups/secret.tar"},
		"network":{"status":"unknown"},
		"software":{"status":"ok","version":"0.2.0-test","commit":"test-commit","channel":"development"},
		"runtime":{"status":"ready"},
		"health":{"status":"healthy"},
		"collected_at":"2026-07-14T12:00:00.000000000Z"
	}`)
	if err := checkNoSensitiveOverviewFields(body); err == nil {
		t.Fatal("expected disclosure guard failure for forbidden path field")
	}
}
