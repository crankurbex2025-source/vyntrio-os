# Lizenzsystem

## Primärmodell (Ziel)

Lizenzierung ist an **Geräte- bzw. Server-Tiers** gebunden (z. B. Community / Pro / Enterprise
als Kapazitäts- und Support-Stufen), **nicht** an fragmentierte Per-Feature-Gates einzelner
Module. Ein Tier umfasst das gesamte Produkt innerhalb seiner Kapazitätsgrenzen; optionale
Add-ons sind Ausnahmen und benötigen ein eigenes ADR.

## Editionen (Tier-Namen)
- Community
- Pro
- Enterprise

## Ziele
Gerätegebundene Tier-Validierung ohne Beeinträchtigung der Systemstabilität. Keine
dauerhafte Deaktivierung einzelner Kernmodule über versteckte Feature-Flags.

## Funktionen
- Hardwarebindung
- USB-Lizenzoption (kommerzielles Aktivierungsmedium — **nicht** Install- oder
  Recovery-USB/ISO; siehe `docs/ADR/0006-appliance-install-recovery-media.md`)
- Online Aktivierung
- Offline Aktivierung
- Digitale Signaturen

## Architektur
License Service validiert signierte Lizenzobjekte lokal. Online-Validierung ergänzt Revocation, Seat- und Laufzeitprüfungen. Kernfunktionen müssen bei temporären Portalstörungen weiterarbeiten.
