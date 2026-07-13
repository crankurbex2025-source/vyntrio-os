package identity

// Canonical security audit event types emitted by v1 identity flows.
const (
	AuditEventBootstrapSucceeded = "identity.bootstrap.succeeded"
	AuditEventLoginSucceeded     = "identity.login.succeeded"
	AuditEventLoginFailure       = "identity.login.failure"
	AuditEventLogoutSucceeded    = "identity.logout.succeeded"
)
