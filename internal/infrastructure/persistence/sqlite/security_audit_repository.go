package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	appidentity "github.com/crankurbex2025-source/vyntrio-os/internal/application/identity"
	"github.com/crankurbex2025-source/vyntrio-os/internal/infrastructure/persistence/sqlite/sqlcgen"
)

// SecurityAuditRepository implements appidentity.SecurityAuditStore.
type SecurityAuditRepository struct {
	q *sqlcgen.Queries
}

var _ appidentity.SecurityAuditStore = (*SecurityAuditRepository)(nil)

// NewSecurityAuditRepository creates an audit store backed by the given database.
func NewSecurityAuditRepository(db *sql.DB) *SecurityAuditRepository {
	return &SecurityAuditRepository{q: sqlcgen.New(db)}
}

// AppendSecurityAuditEvent inserts an append-only audit event.
func (r *SecurityAuditRepository) AppendSecurityAuditEvent(ctx context.Context, input appidentity.AppendSecurityAuditEventInput) error {
	if err := r.q.AppendSecurityAuditEvent(ctx, sqlcgen.AppendSecurityAuditEventParams{
		ID:            input.ID,
		ActorUserID:   nullString(string(input.ActorUserID)),
		SubjectUserID: nullString(string(input.SubjectUserID)),
		EventType:     input.EventType,
		Result:        input.Result,
		IpHash:        nullString(input.IPHash),
		UserAgentHash: nullString(input.UserAgentHash),
		MetadataJson:  nullString(input.MetadataJSON),
	}); err != nil {
		return fmt.Errorf("append security audit event: %w", err)
	}
	return nil
}

// ListSecurityAuditEvents returns a bounded cursor page ordered newest first.
func (r *SecurityAuditRepository) ListSecurityAuditEvents(ctx context.Context, input appidentity.ListSecurityAuditEventsInput) ([]appidentity.SecurityAuditEvent, error) {
	limit, err := validateListLimit(input.Limit)
	if err != nil {
		return nil, err
	}

	params := sqlcgen.ListSecurityAuditEventsParams{
		RowLimit: limit,
	}
	if input.After != nil {
		params.AfterOccurredAt = input.After.OccurredAt
		params.AfterID = nullString(input.After.ID)
	}

	rows, err := r.q.ListSecurityAuditEvents(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list security audit events: %w", err)
	}

	out := make([]appidentity.SecurityAuditEvent, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapAuditEvent(row))
	}
	return out, nil
}

func mapAuditEvent(row sqlcgen.SecurityAuditEvent) appidentity.SecurityAuditEvent {
	return appidentity.SecurityAuditEvent{
		ID:            row.ID,
		OccurredAt:    row.OccurredAt,
		ActorUserID:   optionalUserID(stringFromNull(row.ActorUserID)),
		SubjectUserID: optionalUserID(stringFromNull(row.SubjectUserID)),
		EventType:     row.EventType,
		Result:        row.Result,
		IPHash:        stringFromNull(row.IpHash),
		UserAgentHash: stringFromNull(row.UserAgentHash),
		MetadataJSON:  stringFromNull(row.MetadataJson),
	}
}
