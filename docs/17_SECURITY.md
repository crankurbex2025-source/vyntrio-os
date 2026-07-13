# Security

## Pflichtanforderungen
- HTTPS everywhere
- TLS-Härtung
- Firewall standardmäßig aktiv
- sichere Passwörter
- Audit Logs
- signierte Updates
- sichere API

## Sicherheitsprogramm
- Threat Modeling je Kernmodul
- Secret Handling Policy
- Dependency Scanning
- SAST/DAST in CI
- Security Incident Runbooks
- Responsible Disclosure

## Besondere Risiken
- Remote Admin Surface
- Container Escape
- Unsichere Plugin-Erweiterungen
- Manipulierte Update-Pipelines
- Lizenz-Bypass und Token-Leaks

## Identity & Sessions (v1)

Browser-Auth nutzt serverseitige Sessions (`vyntrio_session`, HttpOnly) und CSRF via `X-CSRF-Token` (geliefert nur im Login-JSON). Details: `docs/ADR/0004-identity-and-access.md`, `docs/11_AUTH.md`.
