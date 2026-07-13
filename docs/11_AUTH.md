# Auth

## Ziele
Sichere Authentifizierung und granulare Autorisierung für Einzelserver und kleine Teams.

## Methoden (v1)
- Login mit Benutzername und Passwort.
- Argon2id für Passwort-Hashes.
- **Serverseitige Sessions** mit opaque Session-ID im HttpOnly-Cookie `vyntrio_session` (SQLite speichert nur Hashes).
- **Kein JWT** und **kein Refresh-Token** für Browser-Auth in v1.
- CSRF-Schutz: roher CSRF-Token nur im Login-JSON-Body; Client sendet `X-CSRF-Token` für mutierende Requests (z. B. Logout).
- 2FA als spätere Erweiterung.

## RBAC (v1)
Fünf feste Rollen:

| Rolle | Intent |
|-------|--------|
| **Owner** | Vollzugriff inkl. Owner-only Aktionen |
| **Administrator** | Admin ohne Owner-only Rechte |
| **Operator** | Betrieb ohne Identity-Admin |
| **User** | Standardbenutzer, eingeschränktes Schreiben |
| **ReadOnly** | Nur Lesen |

Owner-only Admin-Settings: Permission `settings:admin:read` schützt `GET /api/v1/settings`; Permission `settings:admin:write` schützt `PATCH /api/v1/settings/instance` (CSRF erforderlich). Die allgemeinen Permissions `settings:read` / `settings:write` bleiben unverändert und werden von diesen Admin-Endpoints nicht verwendet.

## Sicherheitsregeln
- Sessions widerrufbar (Logout).
- Audit Events für sicherheitsrelevante Identity-Aktionen (`identity.login.succeeded`, `identity.login.failure`, `identity.logout.succeeded`, …) und Settings-Mutationen (`settings.instance_display_name.updated`).
- Keine Benutzer-Enumeration bei Login-Fehlern.
- Secret Handling: keine rohen Session-/CSRF-Tokens in Logs, Audit-Metadaten oder DB.

Siehe auch `docs/ADR/0004-identity-and-access.md`.
