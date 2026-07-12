package settings

import "log/slog"

// LoggerWithSystemSettings returns a logger enriched with validated system settings.
// hostname is always attached; timezone is attached only when includeTimezone is true.
func LoggerWithSystemSettings(logger *slog.Logger, sys SystemSettings, includeTimezone bool) *slog.Logger {
	attrs := []any{slog.String("hostname", sys.Hostname)}
	if includeTimezone {
		attrs = append(attrs, slog.String("timezone", sys.Timezone))
	}
	return logger.With(attrs...)
}
