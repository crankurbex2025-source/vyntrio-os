# Backend

## Service-Zuschnitt
Der Backend-Kern wird als modularer Go-Monorepo mit klar getrennten Packages implementiert.

## Packages
- cmd/api
- cmd/worker
- internal/domain
- internal/application
- internal/infrastructure
- internal/interfaces
- pkg/sdk

## Technische Regeln
- Context propagation in allen IO-Pfaden.
- Strukturierte Fehlerobjekte mit Codes.
- Idempotente Admin-APIs soweit möglich.
- Jobs für langlaufende Operationen.
- Kein direkter Shell-Aufruf ohne Adapter-Layer und Input-Härtung.

## Nebenläufigkeit
Worker-Pool für Systemjobs, Queueing für Updates, Snapshot-Erstellung, Imports und VM-/Container-Aktionen.
