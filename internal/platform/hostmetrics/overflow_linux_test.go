//go:build linux

package hostmetrics

import "testing"

func TestKibToBytesSafeRejectsOverflow(t *testing.T) {
	overflowKib := (^uint64(0) / kibToBytes) + 1
	if _, err := kibToBytesSafe(overflowKib); err == nil {
		t.Fatal("kibToBytesSafe() expected overflow error")
	}
}

func TestMultiplyUint64RejectsSmallestDeterministicOverflow(t *testing.T) {
	const b uint64 = 2
	a := ^uint64(0)/b + 1
	got, err := multiplyUint64(a, b)
	if err == nil {
		t.Fatal("multiplyUint64() expected overflow error")
	}
	if got != 0 {
		t.Fatalf("multiplyUint64() = %d, want 0 on overflow", got)
	}
}

func TestMultiplyUint64RejectsMaxTimesTwo(t *testing.T) {
	got, err := multiplyUint64(^uint64(0), 2)
	if err == nil {
		t.Fatal("multiplyUint64() expected overflow error")
	}
	if got != 0 {
		t.Fatalf("multiplyUint64() = %d, want 0 on overflow", got)
	}
}
