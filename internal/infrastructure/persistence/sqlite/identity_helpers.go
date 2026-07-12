package sqlite

import (
	"database/sql"
	"fmt"

	domainidentity "github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
)

const maxListLimit = 100

func validateListLimit(limit int) (int64, error) {
	if limit <= 0 {
		return 0, fmt.Errorf("limit must be positive")
	}
	if limit > maxListLimit {
		return 0, fmt.Errorf("limit exceeds maximum %d", maxListLimit)
	}
	return int64(limit), nil
}

func nullString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func stringFromNull(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

func optionalUserID(value string) domainidentity.UserID {
	if value == "" {
		return ""
	}
	return domainidentity.UserID(value)
}

func parseRole(role string) (domainidentity.Role, error) {
	return domainidentity.ParseRole(role)
}

func boolFromInt(value int64) bool {
	return value != 0
}

func intFromBool(value bool) int64 {
	if value {
		return 1
	}
	return 0
}
