# Auth

## Ziele
Sichere Authentifizierung und granulare Autorisierung für Einzelserver und kleine Teams.

## Methoden
- Login mit Benutzername/E-Mail und Passwort.
- Argon2id für Passwort-Hashes.
- JWT Access Token plus Refresh Token Rotation.
- 2FA als spätere Erweiterung.

## RBAC
- Administrator
- Power User
- User
- Gast

## Sicherheitsregeln
- Account Lockout/Rate Limits.
- Sessions widerrufbar.
- Audit Logs für sensible Aktionen.
- Secret Rotation und sichere Cookie-/Header-Strategie.
