package health_test

import (
	"context"
	"errors"
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/application/health"
)

type mockDB struct {
	err error
}

func (m mockDB) Ping(_ context.Context) error {
	return m.err
}

func TestReadinessAllOK(t *testing.T) {
	r := health.NewReadiness(mockDB{err: nil})
	res := r.Check(context.Background())
	if !res.ProcessOK || !res.DatabaseOK {
		t.Fatalf("result = %+v, want all ok", res)
	}
}

func TestReadinessDatabaseError(t *testing.T) {
	r := health.NewReadiness(mockDB{err: errors.New("connection refused")})
	res := r.Check(context.Background())
	if !res.ProcessOK || res.DatabaseOK {
		t.Fatalf("result = %+v, want database not ok", res)
	}
}

func TestReadinessNilDB(t *testing.T) {
	r := health.NewReadiness(nil)
	res := r.Check(context.Background())
	if res.DatabaseOK {
		t.Fatal("expected database not ok with nil checker")
	}
}
