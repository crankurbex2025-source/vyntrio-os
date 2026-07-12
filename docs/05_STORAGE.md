# Storage

## Ziele
Ein sicheres und verständliches Storage-Modell mit Fokus auf Privat- und Small-Business-Use-Cases.

## Komponenten
- Disk Manager
- Pool Manager
- Share Manager
- Cache Manager
- Health Monitor
- Backup Manager

## Kernfunktionen
- Laufwerkerkennung inklusive Modell, Seriennummer, Größe und Zustand.
- SMART-Auswertung und proaktive Warnungen.
- Pool-Erstellung auf Basis unterstützter Dateisysteme.
- Shares mit ACL- und Protokollzuweisung.
- Snapshots und Integritätsprüfungen.
- Restore-Workflows mit Validierung und Dry-Run.

## Version-1-Regeln
- EXT4, XFS und BTRFS als Primary Support.
- ZFS erst als spätere Integration hinter Feature Flag.
- Jeder destruktive Storage-Job benötigt Bestätigung und Audit Log.

## APIs
Disk-, Pool-, Share-, Snapshot-, Health- und Restore-Endpunkte mit asynchronem Job-Tracking.
