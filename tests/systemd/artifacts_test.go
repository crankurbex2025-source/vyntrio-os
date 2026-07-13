package systemd_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func artifactPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("..", "..", "distro", "systemd", name)
}

func readArtifact(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(artifactPath(t, name))
	if err != nil {
		t.Fatalf("ReadFile(%s) error: %v", name, err)
	}
	return string(data)
}

func TestServiceUnitRunsAsVyntrioIdentity(t *testing.T) {
	body := readArtifact(t, "vyntrio-api.service")
	for _, want := range []string{
		"User=vyntrio",
		"Group=vyntrio",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("service unit missing %q", want)
		}
	}
	if strings.Contains(body, "DynamicUser=") {
		t.Fatal("service unit must not use DynamicUser")
	}
}

func TestServiceUnitUsesExplicitConfigFlag(t *testing.T) {
	body := readArtifact(t, "vyntrio-api.service")
	want := "ExecStart=/usr/bin/vyntrio-api --config /etc/vyntrio/config.toml"
	if !strings.Contains(body, want) {
		t.Fatalf("ExecStart = %q, want substring %q", body, want)
	}
	if strings.Contains(body, "WorkingDirectory=") {
		t.Fatal("service unit must not set WorkingDirectory; API uses absolute paths only")
	}
}

func TestServiceUnitStateDirectoryMatchesProductionContract(t *testing.T) {
	body := readArtifact(t, "vyntrio-api.service")
	for _, want := range []string{
		"StateDirectory=vyntrio",
		"StateDirectoryMode=0750",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("service unit missing %q", want)
		}
	}
}

func TestServiceUnitConservativeRestartAndGracefulShutdown(t *testing.T) {
	body := readArtifact(t, "vyntrio-api.service")
	for _, want := range []string{
		"Type=simple",
		"Restart=on-failure",
		"RestartSec=5",
		"KillSignal=SIGTERM",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("service unit missing %q", want)
		}
	}
}

func TestServiceUnitSandboxingDirectivesPresent(t *testing.T) {
	body := readArtifact(t, "vyntrio-api.service")
	for _, want := range []string{
		"NoNewPrivileges=yes",
		"CapabilityBoundingSet=",
		"ProtectSystem=strict",
		"ProtectHome=yes",
		"PrivateTmp=yes",
		"PrivateDevices=yes",
		"RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("service unit missing %q", want)
		}
	}
}

func TestSysusersDeclaresStaticVyntrioAccount(t *testing.T) {
	body := readArtifact(t, "vyntrio.sysusers")
	for _, want := range []string{
		`u vyntrio`,
		`m vyntrio vyntrio`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("sysusers missing %q", want)
		}
	}
	if strings.Contains(body, "g vyntrio vyntrio") {
		t.Fatal("sysusers must not use group name as GID field")
	}
	if !strings.Contains(body, "g vyntrio -") {
		t.Fatalf("sysusers missing exact group declaration %q", "g vyntrio -")
	}
	if err := assertSysusersGroupGIDField(body); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(body, "/home/") {
		t.Fatal("sysusers must not assign a home directory")
	}
}

func assertSysusersGroupGIDField(body string) error {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "g" {
			continue
		}
		if fields[1] != "vyntrio" {
			continue
		}
		if len(fields) < 3 {
			return fmt.Errorf("group directive missing GID field: %q", line)
		}
		gid := fields[2]
		if gid == "vyntrio" {
			return fmt.Errorf("group GID field must not repeat group name: %q", line)
		}
		if gid != "-" {
			for _, r := range gid {
				if r < '0' || r > '9' {
					return fmt.Errorf("group GID field must be - or numeric: %q", line)
				}
			}
		}
		return nil
	}
	return fmt.Errorf("group directive for vyntrio not found")
}

func TestTmpfilesConfigDirectoryPermissions(t *testing.T) {
	body := readArtifact(t, "vyntrio.tmpfiles.conf")
	if !strings.Contains(body, "d /etc/vyntrio 0750 root vyntrio") {
		t.Fatalf("tmpfiles missing /etc/vyntrio layout: %q", body)
	}
}

func TestSystemdAnalyzeVerifyProductionUnitDirect(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("systemd-analyze verification requires Linux")
	}
	if _, err := exec.LookPath("systemd-analyze"); err != nil {
		t.Skip("systemd-analyze not available")
	}

	servicePath, err := filepath.Abs(artifactPath(t, "vyntrio-api.service"))
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("systemd-analyze", "verify", servicePath)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return
	}

	if _, statErr := os.Stat("/usr/bin/vyntrio-api"); statErr == nil {
		t.Fatalf("systemd-analyze verify failed with production binary present: %v\n%s", err, out)
	}

	msg := string(out)
	if !strings.Contains(msg, "/usr/bin/vyntrio-api") {
		t.Fatalf("unexpected verify output for missing production binary: %v\n%s", err, out)
	}
}

func TestSystemdAnalyzeVerifyServiceUnitSyntaxCopy(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("systemd-analyze verification requires Linux")
	}
	if _, err := exec.LookPath("systemd-analyze"); err != nil {
		t.Skip("systemd-analyze not available")
	}

	// Separate from TestSystemdAnalyzeVerifyProductionUnitDirect: validate unit
	// syntax and sandbox directives with ExecStart=/bin/true when the production
	// binary is not installed. Production ExecStart is asserted on the real
	// artifact in TestServiceUnitUsesExplicitConfigFlag.
	body := readArtifact(t, "vyntrio-api.service")
	body = strings.Replace(
		body,
		"ExecStart=/usr/bin/vyntrio-api --config /etc/vyntrio/config.toml",
		"ExecStart=/bin/true",
		1,
	)

	tmp := t.TempDir()
	servicePath := filepath.Join(tmp, "vyntrio-api.service")
	if err := os.WriteFile(servicePath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("systemd-analyze", "verify", servicePath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("systemd-analyze verify failed: %v\n%s", err, out)
	}
}
