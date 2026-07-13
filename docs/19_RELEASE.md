# Release

## Produktions-Build (v1, implementiert)

Das Release-Artefakt der API ist **ein einzelnes Go-Binary**, das die
produktive Frontend-UI über `go:embed` enthält.

### Voraussetzungen

- Go **1.24.0** (`go.mod`)
- Node.js **>= 20** (`frontend/package.json`, `engines`) — erforderlich für
  jeden Build, der die produktive UI einbettet

### Standard-Build-Pfad

```bash
make build
```

`make build` hat `ui-stage` als explizite Voraussetzung. Ablauf:

1. `cd frontend && npm run build` — produktiver Vite-Build nach `frontend/dist`.
2. Nur das dedizierte Staging-Verzeichnis
   `internal/interfaces/http/ui/dist/` wird geleert und neu befüllt
   (Kopie von `frontend/dist/.`); Staging bricht ab, wenn `index.html`
   oder Assets fehlen.
3. `go build` kompiliert die Binaries; `go:embed` bettet die gestageten
   Dateien ein. Fehlt das Staging-Verzeichnis, schlägt die Go-Kompilierung
   mit einem klaren Fehler fehl (kein leises leeres UI).

Auch `make test-go` und `make lint` hängen von `ui-stage` ab, sodass die
dokumentierten Kommandos nie ein veraltetes gestagetes UI verwenden. Ein
direkter `go build`-Aufruf ohne vorheriges `make ui-stage` baut das Frontend
**nicht** neu.

`internal/interfaces/http/ui/dist/` ist generiert, per globaler
`dist/`-Regel in `.gitignore` ignoriert und darf **nicht** committet werden.

### Laufzeit-Unabhängigkeit

Das gebaute Binary benötigt zur Laufzeit weder den Repository-Checkout noch
`frontend/dist`, `node_modules`, Vite, einen Dev-Server oder ein
Static-Files-Verzeichnis. Alle UI-Dateien sind im Binary enthalten.

CI (`.github/workflows/ci.yml`) und Release-Validierung
(`.github/workflows/release.yml`) installieren Node, führen `npm ci` im
Frontend aus und rufen `make ui-stage` auf, bevor Go-Kommandos laufen.

### Verifikation

```bash
go test ./...
make verify
make build
cd frontend && npm run typecheck
cd frontend && npm run lint
cd frontend && npm run test:run
cd frontend && npm run build
git diff --check
git status --short
test -z "$(gofmt -l .)"   # gofmt-Gate, identisch zur CI
```

### Betriebshinweise (Troubleshooting)

- Schlägt der Build fehl, weil das eingebettete UI-Input fehlt
  (`pattern all:dist: no matching files found`): den dokumentierten
  Build-Pfad (`make build` bzw. `make ui-stage`) ausführen — niemals
  generierte `dist`-Dateien manuell committen.
- Wirkt ein UI-Asset nach einem Deployment veraltet: `index.html` wird mit
  `Cache-Control: no-cache` bei jedem Laden revalidiert und referenziert
  content-gehashte Assets; die Assets selbst sind unveränderlich gecacht
  (`immutable`). Ein neues Deployment liefert neue Hash-Dateinamen über das
  revalidierte HTML aus.
- Dass API-artige Pfade (`/api/...`) JSON-Fehler statt UI-HTML liefern, ist
  beabsichtigt (API-Isolation, siehe `docs/09_API.md`).

### Betriebs- und Upgrade-Vertrag (Block 7, beschlossen — noch nicht implementiert)

Das Laufzeit-, Dateisystem-, systemd-, Backup-/Restore- und Upgrade-Modell für
den Appliance-Betrieb ist in `docs/ADR/0005-appliance-runtime-operations.md`
festgelegt. Kernpunkte für Releases: unveränderliches Binary getrennt vom
persistenten State (`/var/lib/vyntrio/`), Host-Konfiguration unter
`/etc/vyntrio/`, Konfigurationsänderungen nur per kontrolliertem Neustart,
Backup/Restore als zukünftige lokale Admin-CLI-Operationen (SQLite-konsistent,
nie Raw-Copy einer laufenden Datenbank, nie über Web/API) und
Kompatibilitäts-/Backup-vor-Migration-Policy in späteren Upgrade-Slices.
Docker/OCI-, ISO- und Container-Pfade existieren noch nicht.

## Release-Arten
- Nightly / Development
- Beta
- Stable
- Hotfix

## Prozess
1. Scope einfrieren.
2. Release Notes vorbereiten.
3. Upgrade-Matrix testen.
4. Signaturen erzeugen.
5. Artefakte veröffentlichen.
6. Monitoring und Rollout überwachen.

## Artefakte
- ISO
- Checksums
- Signaturen
- API Specs
- Changelog
- Known Issues
