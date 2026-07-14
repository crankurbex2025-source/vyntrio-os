package hostmetrics_test

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/platform/hostmetrics"
)

var errTestUnavailable = errors.New("unavailable")

func TestParseLoadAvg1mSuccess(t *testing.T) {
	got, err := hostmetrics.ParseLoadAvg1m("0.42 0.30 0.18 2/341 9912\n")
	if err != nil {
		t.Fatalf("ParseLoadAvg1m() error: %v", err)
	}
	if got != 0.42 {
		t.Fatalf("load_1m = %v, want 0.42", got)
	}
}

func TestParseLoadAvg1mRejectsInvalidInput(t *testing.T) {
	cases := []string{
		"",
		strings.Repeat("a", 300),
		"bad",
		"-1.0 0.1 0.1 1/1 1",
		"NaN 0.1 0.1 1/1 1",
	}
	for _, input := range cases {
		if _, err := hostmetrics.ParseLoadAvg1m(input); err == nil {
			t.Fatalf("ParseLoadAvg1m(%q) expected error", input)
		}
	}
}

func TestParseMeminfoSuccessAndUsedDerivation(t *testing.T) {
	content := "MemTotal:       16384 kB\nMemAvailable:    8192 kB\n"
	total, available, err := hostmetrics.ParseMeminfo(content)
	if err != nil {
		t.Fatalf("ParseMeminfo() error: %v", err)
	}
	if total != 16384*1024 {
		t.Fatalf("total = %d", total)
	}
	if available != 8192*1024 {
		t.Fatalf("available = %d", available)
	}
	used, err := hostmetrics.DeriveUsedBytes(total, available)
	if err != nil {
		t.Fatalf("DeriveUsedBytes() error: %v", err)
	}
	if used != total-available {
		t.Fatalf("used = %d, want %d", used, total-available)
	}
}

func TestParseMeminfoRejectsInvalidInput(t *testing.T) {
	cases := []string{
		"",
		strings.Repeat("x\n", 600),
		"MemTotal: 0 kB\nMemAvailable: 0 kB\n",
		"MemTotal: 100 kB\n",
		"MemAvailable: 50 kB\n",
		"MemTotal: 100 kB\nMemAvailable: 200 kB\n",
		"MemTotal: bad kB\nMemAvailable: 50 kB\n",
	}
	for _, input := range cases {
		if _, _, err := hostmetrics.ParseMeminfo(input); err == nil {
			t.Fatalf("ParseMeminfo(%q) expected error", input)
		}
	}
}

func TestMapFSTypeAllowlist(t *testing.T) {
	cases := map[uint32]string{
		0xEF53:     "ext4",
		0x58465342: "xfs",
		0x9123683E: "btrfs",
		0x01021994: "tmpfs",
		0x00000000: "other",
	}
	for magic, want := range cases {
		if got := hostmetrics.MapFSType(magic); got != want {
			t.Fatalf("MapFSType(%#x) = %q, want %q", magic, got, want)
		}
	}
}

func TestCollectorSuccessfulFixtures(t *testing.T) {
	collector := hostmetrics.NewCollector("/var/lib/vyntrio", hostmetrics.CollectorDeps{
		LoadAvg: stubLoadAvgReader{"0.15 0.10 0.05 1/1 1\n"},
		MemInfo: stubMemInfoReader{"MemTotal:       2048 kB\nMemAvailable:    1024 kB\n"},
		StatFS: stubStatFSReader{result: hostmetrics.StatFSResult{
			TotalBytes:     1000,
			AvailableBytes: 400,
			UsedBytes:      600,
			FSType:         "ext4",
		}},
		LogicalCores: func() int { return 4 },
	})

	got := collector.Collect(t.Context())
	if got.CPU.Status != hostmetrics.StatusOK || got.CPU.LogicalCores == nil || *got.CPU.LogicalCores != 4 {
		t.Fatalf("cpu = %+v", got.CPU)
	}
	if got.CPU.Load1m == nil || math.Abs(*got.CPU.Load1m-0.15) > 0.0001 {
		t.Fatalf("cpu load = %+v", got.CPU.Load1m)
	}
	if got.Memory.Status != hostmetrics.StatusOK || got.Memory.UsedBytes == nil || *got.Memory.UsedBytes != 1024*1024 {
		t.Fatalf("memory = %+v", got.Memory)
	}
	if len(got.Filesystems) != 1 || got.Filesystems[0].ID != "state" || got.Filesystems[0].Status != hostmetrics.StatusOK {
		t.Fatalf("filesystems = %+v", got.Filesystems)
	}
}

func TestCollectorIndividualFailuresStayUnavailable(t *testing.T) {
	okDeps := hostmetrics.CollectorDeps{
		LoadAvg: stubLoadAvgReader{"0.15 0.10 0.05 1/1 1\n"},
		MemInfo: stubMemInfoReader{"MemTotal:       2048 kB\nMemAvailable:    1024 kB\n"},
		StatFS: stubStatFSReader{result: hostmetrics.StatFSResult{
			TotalBytes: 1000, AvailableBytes: 400, UsedBytes: 600, FSType: "ext4",
		}},
		LogicalCores: func() int { return 2 },
	}

	cpuOnly := hostmetrics.NewCollector("/var/lib/vyntrio", hostmetrics.CollectorDeps{
		LoadAvg:      failingLoadAvgReader{},
		MemInfo:      okDeps.MemInfo,
		StatFS:       okDeps.StatFS,
		LogicalCores: okDeps.LogicalCores,
	})
	gotCPU := cpuOnly.Collect(t.Context())
	if gotCPU.CPU.Status != hostmetrics.StatusUnavailable || gotCPU.CPU.LogicalCores != nil || gotCPU.CPU.Load1m != nil {
		t.Fatalf("cpu failure = %+v", gotCPU.CPU)
	}
	if gotCPU.Memory.Status != hostmetrics.StatusOK {
		t.Fatalf("memory should remain ok: %+v", gotCPU.Memory)
	}

	memOnly := hostmetrics.NewCollector("/var/lib/vyntrio", hostmetrics.CollectorDeps{
		LoadAvg:      okDeps.LoadAvg,
		MemInfo:      failingMemInfoReader{},
		StatFS:       okDeps.StatFS,
		LogicalCores: okDeps.LogicalCores,
	})
	gotMem := memOnly.Collect(t.Context())
	if gotMem.Memory.Status != hostmetrics.StatusUnavailable || gotMem.Memory.TotalBytes != nil {
		t.Fatalf("memory failure = %+v", gotMem.Memory)
	}

	fsOnly := hostmetrics.NewCollector("/var/lib/vyntrio", hostmetrics.CollectorDeps{
		LoadAvg:      okDeps.LoadAvg,
		MemInfo:      okDeps.MemInfo,
		StatFS:       failingStatFSReader{},
		LogicalCores: okDeps.LogicalCores,
	})
	gotFS := fsOnly.Collect(t.Context())
	if gotFS.Filesystems[0].Status != hostmetrics.StatusUnavailable || gotFS.Filesystems[0].TotalBytes != nil {
		t.Fatalf("filesystem failure = %+v", gotFS.Filesystems[0])
	}
}

func TestCollectorRejectsZeroLogicalCores(t *testing.T) {
	assertInvalidLogicalCoreCount(t, 0)
}

func TestCollectorRejectsNegativeLogicalCores(t *testing.T) {
	assertInvalidLogicalCoreCount(t, -1)
}

func assertInvalidLogicalCoreCount(t *testing.T, cores int) {
	t.Helper()

	okDeps := hostmetrics.CollectorDeps{
		LoadAvg: stubLoadAvgReader{"0.15 0.10 0.05 1/1 1\n"},
		MemInfo: stubMemInfoReader{"MemTotal:       2048 kB\nMemAvailable:    1024 kB\n"},
		StatFS: stubStatFSReader{result: hostmetrics.StatFSResult{
			TotalBytes: 1000, AvailableBytes: 400, UsedBytes: 600, FSType: "ext4",
		}},
		LogicalCores: func() int { return cores },
	}
	got := hostmetrics.NewCollector("/var/lib/vyntrio", okDeps).Collect(t.Context())
	if got.CPU.Status != hostmetrics.StatusUnavailable || got.CPU.LogicalCores != nil || got.CPU.Load1m != nil {
		t.Fatalf("cpu = %+v, want unavailable without numeric fields", got.CPU)
	}
	if got.Memory.Status != hostmetrics.StatusOK || got.Memory.TotalBytes == nil {
		t.Fatalf("memory = %+v, want ok with numeric fields", got.Memory)
	}
	if got.Filesystems[0].Status != hostmetrics.StatusOK || got.Filesystems[0].TotalBytes == nil {
		t.Fatalf("filesystem = %+v, want ok with numeric fields", got.Filesystems[0])
	}
}

func TestParseMeminfoRejectsKibToBytesOverflow(t *testing.T) {
	overflowKib := (^uint64(0) / 1024) + 1
	content := fmt.Sprintf("MemTotal: %d kB\nMemAvailable: 1 kB\n", overflowKib)
	if _, _, err := hostmetrics.ParseMeminfo(content); err == nil {
		t.Fatal("ParseMeminfo() expected overflow error")
	}
}

func TestCollectorMemoryUnavailableOnMeminfoOverflow(t *testing.T) {
	overflowKib := (^uint64(0) / 1024) + 1
	content := fmt.Sprintf("MemTotal: %d kB\nMemAvailable: 1 kB\n", overflowKib)
	collector := hostmetrics.NewCollector("/var/lib/vyntrio", hostmetrics.CollectorDeps{
		LoadAvg: stubLoadAvgReader{"0.15 0.10 0.05 1/1 1\n"},
		MemInfo: stubMemInfoReader{content},
		StatFS: stubStatFSReader{result: hostmetrics.StatFSResult{
			TotalBytes: 1000, AvailableBytes: 400, UsedBytes: 600, FSType: "ext4",
		}},
		LogicalCores: func() int { return 2 },
	})

	got := collector.Collect(t.Context())
	if got.Memory.Status != hostmetrics.StatusUnavailable || got.Memory.TotalBytes != nil {
		t.Fatalf("memory = %+v, want unavailable without numeric fields", got.Memory)
	}
	if got.CPU.Status != hostmetrics.StatusOK {
		t.Fatalf("cpu = %+v, want ok", got.CPU)
	}
	if got.Filesystems[0].Status != hostmetrics.StatusOK {
		t.Fatalf("filesystem = %+v, want ok", got.Filesystems[0])
	}
}

func TestStatFSReaderUsesOnlyProvidedStateDir(t *testing.T) {
	reader := recordingStatFSReader{}
	_ = hostmetrics.NewCollector("/var/lib/vyntrio", hostmetrics.CollectorDeps{
		LoadAvg:      stubLoadAvgReader{"0.1 0.1 0.1 1/1 1\n"},
		MemInfo:      stubMemInfoReader{"MemTotal: 1024 kB\nMemAvailable: 512 kB\n"},
		StatFS:       &reader,
		LogicalCores: func() int { return 1 },
	}).Collect(t.Context())
	if reader.got != "/var/lib/vyntrio" {
		t.Fatalf("StatStateFilesystem path = %q", reader.got)
	}
	if reader.got == "/etc/vyntrio" || reader.got == "/var/lib/vyntrio/backups" {
		t.Fatal("statfs used forbidden path")
	}
}

type stubLoadAvgReader struct {
	content string
}

func (s stubLoadAvgReader) ReadLoadAvg() (string, error) { return s.content, nil }

type failingLoadAvgReader struct{}

func (failingLoadAvgReader) ReadLoadAvg() (string, error) { return "", errTestUnavailable }

type stubMemInfoReader struct {
	content string
}

func (s stubMemInfoReader) ReadMemInfo() (string, error) { return s.content, nil }

type failingMemInfoReader struct{}

func (failingMemInfoReader) ReadMemInfo() (string, error) { return "", errTestUnavailable }

type stubStatFSReader struct {
	result hostmetrics.StatFSResult
}

func (s stubStatFSReader) StatStateFilesystem(_ string) (hostmetrics.StatFSResult, error) {
	return s.result, nil
}

type failingStatFSReader struct{}

func (failingStatFSReader) StatStateFilesystem(_ string) (hostmetrics.StatFSResult, error) {
	return hostmetrics.StatFSResult{}, errTestUnavailable
}

type recordingStatFSReader struct {
	got string
}

func (r *recordingStatFSReader) StatStateFilesystem(stateDir string) (hostmetrics.StatFSResult, error) {
	r.got = stateDir
	return hostmetrics.StatFSResult{
		TotalBytes: 100, AvailableBytes: 50, UsedBytes: 50, FSType: "ext4",
	}, nil
}
