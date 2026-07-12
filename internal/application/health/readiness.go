package health

import "context"

// DatabaseChecker checks database connectivity (implemented by persistence layer).
type DatabaseChecker interface {
	Ping(ctx context.Context) error
}

// Readiness evaluates process and dependency readiness for /readyz.
type Readiness struct {
	db DatabaseChecker
}

// NewReadiness creates a readiness evaluator.
func NewReadiness(db DatabaseChecker) *Readiness {
	return &Readiness{db: db}
}

// Result holds readiness check outcomes.
type Result struct {
	ProcessOK  bool
	DatabaseOK bool
}

// Check runs readiness probes.
func (r *Readiness) Check(ctx context.Context) Result {
	res := Result{ProcessOK: true}
	if r.db == nil {
		res.DatabaseOK = false
		return res
	}
	if err := r.db.Ping(ctx); err != nil {
		res.DatabaseOK = false
		return res
	}
	res.DatabaseOK = true
	return res
}
