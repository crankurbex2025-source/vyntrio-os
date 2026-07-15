# Lizenzsystem

## Editionen
- Community
- Pro
- Enterprise

## Ziele
Edition-basierte Feature-Freischaltung ohne Beeinträchtigung der Systemstabilität.

## Funktionen
- Hardwarebindung
- USB-Lizenzoption (kommerzielles Aktivierungsmedium — **nicht** Install- oder
  Recovery-USB/ISO; siehe `docs/ADR/0006-appliance-install-recovery-media.md`)
- Online Aktivierung
- Offline Aktivierung
- Digitale Signaturen

## Architektur
License Service validiert signierte Lizenzobjekte lokal. Online-Validierung ergänzt Revocation, Seat- und Laufzeitprüfungen. Kernfunktionen müssen bei temporären Portalstörungen weiterarbeiten.
