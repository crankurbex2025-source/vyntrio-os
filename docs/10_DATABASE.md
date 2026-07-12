# Datenbank

## Initiale Wahl
SQLite als eingebettete Appliance-Datenbank für Installationseinfachheit und geringe Betriebsanforderungen.

## Spätere Wahl
PostgreSQL für skalierte Installationen, Replikations- oder Enterprise-Szenarien.

## Kernentitäten
Users, Roles, Permissions, Systems, Disks, Pools, Shares, Containers, VirtualMachines, Licenses, Updates, Logs, Notifications, Settings.

## Datenbankregeln
- Migrations sind versioniert und reversibel, soweit fachlich möglich.
- Soft Delete nur dort, wo Audit oder Wiederherstellung relevant ist.
- Keine JSON-Felder als Ersatz für sauberes Modell, außer bei klaren Extensibility-Cases.
- Indizes anhand echter Query-Pfade definieren.
