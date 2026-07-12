# 17 — Security

> **Status:** DRAFT — security requirements must be defined before production use.

## Security principles (proposed)

1. **Least privilege** — services and humans get minimum required access
2. **No secrets in git** — use environment variables and secret managers
3. **Secure defaults** — fail closed, deny by default
4. **Auditability** — security-relevant actions are logged
5. **Dependency hygiene** — track and patch third-party vulnerabilities

## Secret handling

- Never commit `.env`, keys, tokens, or credentials
- Use `.env.example` for non-secret variable names only
- Rotate credentials exposed in chat, logs, or tickets immediately
- TODO: Define production secret store (e.g. GitHub Secrets, Vault, cloud KMS)

## Authentication and authorization

TODO:

- User identity model
- Session vs token strategy
- Role/permission model
- Service-to-service auth

## Data protection

TODO:

- Data classification
- Encryption at rest and in transit
- PII handling and retention
- Backup and recovery security

## Application security

TODO:

- Input validation standards
- OWASP ASVS or equivalent target level
- CSRF/XSS/SSRF considerations for frontend
- Rate limiting and abuse prevention

## Supply chain

- Pin dependencies where practical
- Enable Dependabot or equivalent (TODO: Phase 1+)
- Review new dependencies via ADR for critical paths

## CI/CD security

Current (Phase 0):

- No secrets in workflow files
- Foundation checks only

TODO:

- Branch protection rules on `main`
- Required status checks
- CODEOWNERS for sensitive paths

## Incident response

TODO:

- Reporting channel
- Severity levels
- Patch/release process

## Related documents

- [02_ARCHITECTURE.md](02_ARCHITECTURE.md)
- [18_TESTING.md](18_TESTING.md)
- [21_CURSOR_REPO_SETUP.md](21_CURSOR_REPO_SETUP.md)
