# Updates

## Ziele
Sichere, nachvollziehbare und rollback-fähige Produktupdates.

## Funktionen
- Signierte Updates
- Versionierung
- Rollback
- Sicherheitsupdates
- Release Channels: Development, Beta, Stable

## Architektur
Ein Update Agent prüft Kanäle, lädt signierte Artefakte, validiert Metadaten, führt Preflight-Checks aus und erstellt einen Rollback-Pfad.

## Regeln
- Kein unsigniertes Paket.
- Channel-Wechsel auditieren.
- Kompatibilitätsmatrix zwischen Versionen pflegen.
