package identity

import "time"

// FormatUTCTime formats timestamps for SQLite TEXT columns.
func FormatUTCTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
