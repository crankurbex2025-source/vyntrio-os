# Storage

## Ziele
Ein sicheres und verständliches Storage-Modell mit Fokus auf Privat- und Small-Business-Use-Cases.

## Implementiert (Block 12 / Slice 12.1)

**Read-only Storage-Inventar** — erste echte Appliance-Grundlage ohne destruktive Operationen.

### Paketgrenzen
| Schicht | Paket | Verantwortung |
|---------|-------|---------------|
| Platform | `internal/platform/storageinventory` | sysfs/mountinfo-Discovery, Eligibility-Klassifikation |
| Application | `internal/application/storage` | API-DTO-Mapping |
| Interface | `GET /api/v1/storage/disks` | Session + `storage:read` |

### Discovery (Linux)
- Blockgeräte unter `/sys/block` (loop, ram, dm-, zram, fd ausgeschlossen)
- Mount-Kontext aus `/proc/self/mountinfo` (Root- und State-Filesystem-Zuordnung)
- Keine Kernel-Namen (`/dev/sdX`) in der API — nur stabile opaque IDs (`disk-<sha256-prefix>`)

### Eligibility-Regeln (fail-closed)
| Reason | Bedeutung |
|--------|-----------|
| `root_disk` | Gerät trägt `/` |
| `state_filesystem` | Gerät trägt `state_dir` (`/var/lib/vyntrio/`) |
| `removable` | sysfs `removable=1` |
| `read_only` | sysfs `ro=1` |
| `mounted_in_use` | gemountet, nicht Root/State |
| `unsupported_filesystem` | FS nicht in {ext4, xfs, btrfs} |
| `install_media` | optisches Laufwerk (`sr*`) |
| `virtual_device` | als virtuell markiert |
| `ambiguous_identity` | fehlende Größe, leere Identität, Integritätsfehler |

**Eligible** nur wenn keine Ausschlussgründe und Größe bekannt > 0.

### Unterstützte Dateisysteme (v1 Kandidaten)
- ext4, xfs, btrfs (ext2/ext3 werden als ext4 normalisiert)

### API
`GET /api/v1/storage/disks` — siehe `docs/09_API.md`.

### Sicherheit
- **Keine Mutation** in diesem Slice (kein Format, kein Wipe, kein Pool)
- **Kein Auto-Select** eines Zielgeräts
- **Fail-closed** bei mehrdeutiger Identität
- Install-Medium, Root- und State-Disk immer ausgeschlossen

## Implementiert (Stage 3 / Slice 11.1 — Storage UI foundation)

**Read-only Storage-Übersicht in der lokalen WebGUI** — erste ehrliche NAS/Storage-Oberfläche
im Appliance-Dashboard (keine Mutation, keine Pools/Shares).

### Frontend
| Komponente | Pfad |
|------------|------|
| DTO-Parser | `frontend/src/features/storage/storageDto.ts` |
| Storage-Shell | `frontend/src/features/storage/StorageShell.tsx` |
| Navigation | `ApplianceApp` → „Storage inventory“ von der Overview |
| API-Aufruf | `GET /api/v1/storage/disks` (Session + `storage:read`) |

### Verhalten
- Zeigt klassifizierte Blockgeräte mit Eligibility und Ausschlussgründen
- Expliziter Hinweis: Pools, Shares, SMART und destruktive Aktionen **nicht verfügbar**
- Kein Polling; Daten werden beim Öffnen der Storage-Ansicht geladen
- 403 → ehrliche Meldung auf der Overview (kein Zugriff)

## Implementiert (Slice 12.2 — Storage layout foundation)

**Read-only Pools/Shares-Layout** — ehrliche leere Zustände bis Mutation-Slices
existieren.

### API
| Endpoint | Verhalten |
|----------|-----------|
| `GET /api/v1/storage/pools` | `pools: []`, `pool_management: "not_available"`, `mutation_available: false` |
| `GET /api/v1/storage/shares` | `shares: []`, `share_management` / `protocol_support: "not_available"` |
| `GET /api/v1/overview` → `storage` | Aggregierte Zähler: Disk-Inventar + deklarierte Pools / geplante Shares (`pool_count`, `share_count` aus `{state_dir}/storage/pools.json`) |

### Application
| Paket | Datei | Verantwortung |
|-------|-------|---------------|
| Application | `pools.go`, `shares.go`, `summary.go` | Layout-Read-Models |
| Interface | `handlers/storage.go` | `ServePools`, `ServeShares` |
| Frontend | `StorageShell.tsx`, `OverviewShell.tsx` | Layout-Status-Panels |

### Verhalten
- Pools- und Shares-Endpunkte spiegeln Inventar-Gesundheit (`inventory_status`)
- UI zeigt „0 configured“ und explizit „not available“ — **keine** Pool-/Share-Erstellung
- Overview fasst Storage-Zähler ohne extra API-Roundtrip beim Dashboard-Laden zusammen

## Implementiert (Slice 12.3 — Declared pool / dataset / share foundation)

**Erste echte Storage-Mutation ohne Disk-Format:**

| Endpoint | Verhalten |
|----------|-----------|
| `POST /api/v1/storage/pools` | Declares a pool from eligible opaque disk IDs; persists under `state_dir/storage/pools.json`; **does not format disks** |
| `POST /api/v1/storage/pools/{id}/datasets` | Prepares a dataset plan (`path_intent`) |
| `POST /api/v1/storage/shares` | Prepares a share plan (`protocol: planned`) |
| `GET .../pools` / `.../shares` | Return persisted declared/planned objects |

### Sicherheit
- Nur `eligibility=eligible` Disks; Root/State/Removable bleiben ausgeschlossen
- `confirm: true` erforderlich
- Disks können nicht zwei Pools zugewiesen werden
- UI und API sagen explizit: **kein Format, kein SMB/NFS**

### Platform
`internal/platform/storagepool` — JSON-Store unter dem Appliance-State-Verzeichnis.

## Nicht implementiert (bewusst zurückgestellt)
- RAID, actual disk format/wipe, ZFS/btrfs pool create on disk
- SMB/NFS/iSCSI daemon integration
- Snapshots, Deduplication, Parity
- SMART-Auswertung und Health-Monitor
- Storage-CLI (`vyntrio-storage`)
- ZFS (später hinter Feature Flag)
- Async Job-Tracking für Storage-Operationen

## Geplante Komponenten (Roadmap)
- Disk Manager (erweitert)
- Pool Manager
- Share Manager
- Cache Manager
- Health Monitor
- Backup Manager (NAS-Artefakte — getrennt von `vyntrio-backup` State-Restore)

## Version-1-Regeln (für zukünftige Mutation-Slices)
- EXT4, XFS und BTRFS als Primary Support.
- ZFS erst als spätere Integration hinter Feature Flag.
- Jeder destruktive Storage-Job benötigt Bestätigung und Audit Log.
