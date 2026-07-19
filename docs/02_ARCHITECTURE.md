# Architektur

## Architekturziel
Vyntrio OS verwendet eine modulare Clean Architecture mit klarer Trennung zwischen Domain, Application, Infrastructure und Interfaces. Das System wird als lokales Appliance-OS mit integriertem Control Plane konzipiert.

## Schichten
### Domain
Enthält die Geschäftslogik für Benutzer, Storage, Container, VMs, Netzwerk, Updates, Lizenzen und Plugins. Keine Framework-Abhängigkeiten.

### Application
Orchestriert Use Cases, Policies, Jobs, Validierung, Event-Publishing und Berechtigungsprüfungen.

### Infrastructure
Adapter für SQLite/PostgreSQL, Docker, libvirt, systemd, nftables, SMART, Update-Signing, Dateisystem-Tools und Lizenzsignaturen.

### Interfaces
REST-API, WebSocket-Streams, CLI-Service-Wrappers, Installer-UI und React-Frontend.

## Kernmodule
- Identity & Access
- System Inventory
- Storage Service
- Network Service
- Container Service
- VM Service
- Backup Service
- Notification Service
- License Service
- Update Service
- Marketplace/Plugin Service

## Kommunikationsmodell
- Synchronous: REST für CRUD, Konfiguration und Admin-Aktionen.
- Real-time: WebSocket für Status, Logs, Progress, Alerts.
- Internal: Event Bus für Domain Events und Long-Running Jobs.

## Persistenz
Start mit SQLite für Appliance-Einfachheit; PostgreSQL als spätere Option für größere Installationen und Enterprise-Szenarien.

## Sicherheitsarchitektur
- TLS-gesicherte Oberfläche.
- Argon2 für Passwort-Hashes.
- Signierte Updates und Plugin-Pakete.
- RBAC mit auditierbaren Admin-Aktionen.
- Secrets niemals im Frontend.

## Deployment-Modell
Control Plane läuft lokal auf dem Host. Systemnahe Services werden über systemd verwaltet. API und Worker laufen als getrennte Go-Binaries oder Prozesse.

**Implementiert (v1):** Das produktive Frontend wird als statisches Vite-Bundle gebaut, per `go:embed` in das API-Binary eingebettet und same-origin vom selben Binary ausgeliefert (`/assets/*` plus `index.html`-SPA-Fallback für erlaubte GET/HEAD-Pfade). API-, Health- und Readiness-Routen behalten Priorität. Der API-Server lädt Laufzeitkonfiguration aus `/etc/vyntrio/config.toml` (TOML, Fail-Closed, Neustart statt Live-Reload); persistenter State unter `/var/lib/vyntrio/` mit startup-time Pfad-/Symlink-Validierung (ohne race-freie SQLite-Eindämmung). **Slice 7.3:** systemd-Unit, statisches `vyntrio`-Konto und Ownership-Modell (`distro/systemd/`). Details: `docs/09_API.md`, `docs/17_SECURITY.md`, `docs/19_RELEASE.md`, `docs/ADR/0005-appliance-runtime-operations.md`.

**Beschlossener Block-7-Vertrag (teilweise implementiert):** Backup-Befehl
`vyntrio-backup` (Slice 7.9); `vyntrio-restore` (Slice 7.12);
Restore-Sicherheitsvertrag (`docs/ops/restore-safety-contract.md`, Slice 7.11).
Packaging bleibt zukünftige Slice. Docker/OCI, ISO, Kubernetes, Clustering und
Multi-Node bleiben ausdrücklich zurückgestellt.

**Block 8 / Slice 8.1 (implementiert):** Authentifizierte read-only Appliance-Overview
(`GET /api/v1/overview`, Permission `system:health`) als Control-Center-Landing;
Application-Layer-DTO in `internal/application/overview`; React-Overview-Shell mit
Session-Bootstrap über Overview statt Owner-only Settings. Owner-Settings bleiben
separat unter `settings:admin:read`/`write`. Keine Host-Metriken, Polling oder
WebSockets in 8.1.

**Block 8 / Slice 8.3 (implementiert):** Read-only Host-Metriken im Overview-DTO
(`host.cpu`, `host.memory`, `host.filesystems[id=state]`) via
`internal/platform/hostmetrics` — direkte in-process `/proc`- und statfs-Sammlung
nur auf dem validierten `state_dir`; per-Metrik `unavailable`-Degradation in HTTP 200.
Kein CPU-%, Netzwerk, Backup-Artefakt-Exposure, Mount-Inventory oder Privilege-Expansion.

**Block 8 / Slice 8.5 (implementiert):** Read-only lokaler Backup-Status im Overview-DTO
(`backup.status`, optional `completed_at`, `ever_succeeded`, `failure`) via
`internal/platform/backupstatus` — API liest nur ein root-geschriebenes Status-Sidecar
unter dem validierten `state_dir`; `vyntrio-backup` ist alleiniger Schreiber.
Per-Status `unavailable`-Degradation in HTTP 200. Kein Backup-Trigger, keine Artefakt-Enumeration.

**Block 8 / Slice 8.7 (implementiert):** Read-only Netzwerk-Präsenz im Overview-DTO
(`network.status` ∈ `available` | `unknown` | `unavailable`) via
`internal/platform/netpresence` — in-process `net.Interfaces()` (Linux) nur Flags und
HardwareAddr-Präsenz; keine Schnittstellen-/MAC-/IP-Exposure. Per-Sektion
`unavailable`-Degradation in HTTP 200. Kein Internet-, DNS- oder Reachability-Nachweis.

**Block 8 / Slice 8.9 (implementiert):** Read-only Software-/Release-Status im Overview-DTO
(`software.status`, optional `version`, `commit`, `channel`) — materialisiert nur aus
bereits lokaler Loader-Metadaten (eingebettete Version/Commit, Kanal aus `api.environment`);
kein Update-Check, keine Paket- oder Host-Inventory.

**Block 8 / Slice 8.10 (implementiert):** Read-only Runtime-Readiness im Overview-DTO
(`runtime.status` ∈ `ready` | `degraded` | `unknown`, optional `note: database`) —
abgeleitet aus bestehendem `readiness` und `service` ohne neue Probes; UI nutzt
`collected_at` für Zeitbezug.

**Block 8 / Slice 8.11 (implementiert):** Read-only Health-Summary im Overview-DTO
(`health.status` ∈ `healthy` | `warning` | `unknown`, optional `note: database|backup`) —
abgeleitet aus bestehendem `runtime` und `backup` ohne neue Probes.

**Block 9 / Slice 9.2 (Vertrag, dokumentiert):** Install- und Recovery-Medien
(ADR-0006) — getrennte Boot-USB/ISO-Rollen für Greenfield-Installation vs.
offline Recovery-Umgebung; persistenter State nur auf Zielplatte
(`/var/lib/vyntrio/`, `/etc/vyntrio/` per ADR-0005); kein Owner-Seeding auf
generischem Medium; License-USB (`docs/15_LICENSE.md`) ist separat. Noch keine
ISO-/Installer-Implementierung.

**Block 9 / Slice 9.3 (Gerüst, dokumentiert):** Deklaratives Install-Media-
Gerüst in `distro/install-media/` (`manifest.yaml`, Config-Template, README) —
listet Payloads und Zielpfade; referenziert `distro/systemd/*`; kein Boot/ISO-
Build, keine Partitionierung, kein Installer-Lauf.

**Block 9 / Slice 9.4 (Gerüst, dokumentiert):** Deklaratives Recovery-Media-
Gerüst in `distro/recovery-media/` (`manifest.yaml`, README) — listet Offline-
Recovery-Tooling und Zielplatten-Interaktion; getrennt von Install-Media; kein
Restore-CLI, kein Boot/ISO-Build.

**Block 9 / Slice 9.5 (Vertrag, dokumentiert):** Install-Image-Build-Vertrag in
`distro/install-media/build-contract.md` — definiert Build-Inputs/Outputs,
Ein-/Ausschlüsse und Verweis auf `manifest.yaml`; kein ISO/USB-Generator.

**Block 9 / Slice 9.6 (implementiert):** Lokales Install-Payload-Staging via
`make install-media-stage` — validiert Build-Artefakte, kopiert nur
Manifest-Payloads nach `distro/install-media/staging/payload/`; kein Boot/ISO,
kein Zielplatten-Schreiben.

**Block 9 / Slice 9.7 (Gerüst, dokumentiert):** Live/Boot-Envelope-Gerüst in
`distro/install-media/envelope-contract.md` und `envelope-manifest.yaml` —
definiert `boot/`, `live_root/`, `payload/`-Schichten und Bezug zum lokalen
Staging; deklarativ nur; kein ISO/USB-Generator, kein Boot/Live-Root-Lauf.

**Block 9 / Slice 9.8 (implementiert):** Lokale Install-Envelope-Assembly via
`make install-media-envelope` — konsumiert `staging/payload/`, erzeugt
`distro/install-media/envelope/` mit deklarierten Schichten; validiert nur
Manifest-Payloads; `boot/`/`live_root/` Platzhalter; kein ISO/USB, kein
Zielplatten-Schreiben.

**Block 9 / Slice 9.9 (Gerüst, dokumentiert):** Bootability-Initialisierungs-
Gerüst in `distro/install-media/bootability-contract.md` und
`bootability-manifest.yaml` — definiert bootfähige Initialisierung über
Envelope-Assembly hinaus; Bootability ≠ Installation; deklarativ nur.

**Block 9 / Slice 9.11 (implementiert, Stub):** Bootability-Foundation via
`make install-media-bootability` — ersetzt `boot/`/`live_root/`-Platzhalter durch
Struktur-Stubs, validiert Layout, schreibt `BOOTABILITY.txt`, emittiert
`distro/install-media/build/*-NOT-BOOTABLE.stub.tar` (**nicht** firmware-bootfähig);
USB-Creator und echtes ISO/USB deferred. Vertrag:
`docs/ops/install-media-bootability.md`.

**Block 9 / Slices 9.12–9.17 (implementiert):** USB-first Boot-Media-Kette über
die Foundation hinaus — 9.12 realer Boot-Chain (echter Kernel/Initrd/GRUB-Core +
`grub.cfg`, `make install-media-image`), 9.13 firmware-bootfähiges Raw-BIOS-Image
via `grub-bios-setup` + strukturelle Verifikation (`make install-media-wrap`),
9.14 gated qemu-Runtime-Boot-Harness + First-Boot-Dashboard-Pfad
(`make install-media-runtime`), 9.15 minimale Live-Rootfs-Userland (Loader/libc-
Closure, busybox, `vyntrio-api`, `/init`), verifiziert per chroot-Exec +
Dashboard-HTTP-200-Probe (`make install-media-live-rootfs`), 9.16
Live-Initramfs-Hardware-Enablement — dep-aufgelöste Storage/Netzwerk-Kernelmodule
+ DHCP-Bring-up in `/init`, Live-Initramfs neu emittiert
(`make install-media-hardware`), und 9.17 Image-Initrd-Swap
(`make install-media-initrd-swap`) — das Firmware-Image bootet jetzt das Vyntrio
Live-Initramfs statt des Host-Debian-Initrd, per sha256 aus dem Image bewiesen,
fail-closed mit Host-Initrd-Restore. Damit ist **Stage 1 abgeschlossen**. Ehrlich:
kein verifizierter Runtime-Boot (kein qemu hier), Dashboard-on-Boot und
USB-Creator weiter deferred. Verträge: `distro/install-media/bootability-contract.md`,
`docs/ops/install-media-{image,wrapper,runtime-boot,live-rootfs,hardware-enable,initrd-swap,runtime-verify}.md`.

**Stage 2 / Slice 10.1 (implementiert, hier blockiert):** Runtime-Boot-Verifikation
via `make install-media-runtime-verify` (`scripts/verify-live-boot-runtime.sh`) —
erste Stage-2-Slice (First-Boot-Dashboard), **unabhängig** vom staged-installer
„Block 10 / Slice 10.1“ (Audit) weiter unten. Ein **read-only** Verifier bootet
das initrd-getauschte Image auf einem VM/Hardware-Harness und erbringt einen
**echten HTTP-200-Nachweis vom gebooteten Gast** (qemu user-net hostfwd), oder
**fail-closed** mit exakten Blockern. Er ändert **niemals** Image/Initrd/
Bootloader/live-rootfs (per sha256 vor/nach im Test erzwungen). Auf diesem Host:
kein `qemu`/`/dev/kvm` → `status: blocked`, `chain_modified: false`. Dashboard-on-
Boot-Nachweis = echte HTTP-Antwort, nicht chroot/strukturell.

**Stage 2 / Slice 10.2 (implementiert — lokale Dashboard-Stabilität):** Die lokale
Browser-WebGUI (`vyntrio-api`) ist die **primäre** Verwaltungsfläche (dashboard-first).
Behobener Defekt: `firstboot.sh` lief unter `#!/usr/bin/env bash`, aber die
Live-Initramfs liefert nur busybox (kein bash/`env`) — der Dashboard-Entrypoint
konnte nie starten; jetzt `#!/bin/sh` (busybox/POSIX). Das Dashboard wird
**beaufsichtigt** (begrenzter Respawn-Loop) und bleibt oben; die Build-Zeit-Probe
beweist zusätzlich `/readyz` (DB-ready), nicht nur `/`. **Slice 10.3** ergänzt
eine klare First-Boot-Onboarding-Ausgabe (Boot → lokales Dashboard → Anmeldung/
Owner → read-only Übersicht) und die Landing-Seiten-Sektion `#first-boot-setup`,
die denselben lokalen Flow wiedergibt. **Slice 10.4** entfernt den TLS-**Config-Blocker**:
`vyntrio-api` unterstützt optionale `tls_cert_file`/`tls_key_file` und HTTPS;
Runtime-TLS-Prep via `prepare-live-dashboard-tls.sh` wenn `VYNTRIO_LAN_BIND_IP`
gesetzt (Loopback-HTTP bleibt Default). `tls_readiness: ready`;
`dashboard_reachable: false` ehrlich (kein Boot-Harness). Kein Installer/Target-Write,
Boot/Image-Chain unverändert. `docs/ops/install-media-dashboard-tls.md`.

**Block 10 / Slice 10.1 (Audit, dokumentiert):** Read-only Installer-Vertrags-
Audit — Grenze zwischen Media-Erstellung (Block 9) und Install-Ausführung
(Block 10) aus ADR-0004/0005/0006 abgeleitet; kein Installer-Code.

**Block 10 / Slice 10.2 (Vertrag, dokumentiert):** Formeller Installer-Vertrag
in `docs/ADR/0007-appliance-installer-contract.md` — Preflight, Zielplatten-
Allowlist, Bootstrap-Handoff, Ausschlüsse; kein Installer-Implementierung.

**Block 10 / Slice 10.3 (implementiert):** Read-only Installer-Preflight via
`make installer-preflight` — validiert Install-Media-Kontext und Payload-
Inventar gegen ADR-0007; kein Zielplatten-Schreiben, kein Bootstrap.

**Block 10 / Slice 10.4 (Gerüst, dokumentiert):** Zielplatten-Layout-Gerüst in
`distro/installer/target-layout-manifest.yaml` und `target-layout-contract.md` —
definiert Allowlist-Pfade und leere State-Verzeichnisse; Partitionierung
deferred; read-only Validierung via `make installer-layout-plan`; kein
Zielplatten-Schreiben.

**Block 10 / Slice 10.5 (implementiert):** Preflight-gated Mutation-Dry-Run-Stub via
`make installer-mutation-stub` — erfordert erfolgreiches Preflight, konsumiert
Layout-Plan, schreibt nur lokales `dry-run/MUTATION_STUB.txt`; kein
Zielplatten-Schreiben.

**Block 10 / Slice 10.6 (implementiert):** Erste Ziel-Mutation — leere State-
Verzeichnisse unter `distro/installer/target-sandbox/` via
`make installer-mutate-directories`; preflight-gated; kein Host-`/etc`/`/var/lib`,
kein Payload-Copy.

**Block 10 / Slice 10.7 (implementiert):** Manifest-Payload-Copy nach
`distro/installer/target-sandbox/` via `make installer-copy-payloads`;
preflight-gated; nur sechs Allowlist-Dateien; kein Service-Enablement, kein Bootstrap.

**Block 10 / Slice 10.8 (implementiert):** Service-Enablement-Vorbereitung via
`make installer-prepare-service` — preflight/copy-gated; Prep-Marker und
`SERVICE_PREP.txt` in `target-sandbox/`; kein systemctl, kein Service-Start,
kein Bootstrap.

**Block 10 / Slice 10.9 (implementiert):** Kontrolliertes Service-Enablement via
`make installer-enable-service` — preflight/prep-gated; systemd wants-Symlink und
`SERVICE_ENABLE.txt` in `target-sandbox/`; kein systemctl, kein Service-Start,
kein Bootstrap.

**Block 12 / Slice 12.1 (implementiert):** Read-only Storage-Inventar via
`GET /api/v1/storage/disks` — Linux-sysfs/mountinfo-Discovery in
`internal/platform/storageinventory`, Eligibility-Klassifikation (fail-closed),
opaque Geräte-IDs; Permission `storage:read`; keine Mutation, kein Pool/Share/SMART.
Details: `docs/05_STORAGE.md`.

**Block 9 / Slice 9.10 (implementiert):** Release-Artefakt-Verifikation via
`vyntrio-verify-artifact` — JSON-Manifest `vyntrio-release-manifest-v1`, lokale
SHA-256- und Größenprüfung, fail-closed Parsing; Ed25519-Signaturfeld reserviert
aber nicht verifiziert; keine Disk-Writes, kein Media-Build. Vertrag:
`docs/ops/release-artifact-verification.md`.

**Block 10 / Slice 10.10 (implementiert):** Read-only Install-Target- und
Media-Preflight via `vyntrio-installer preflight` — explizite `--target-disk-id`,
storageinventory-basierte Eignung, optionale Envelope- und Release-Manifest-Prüfung;
keine Partitionierung, kein Schreiben, kein Install. Vertrag:
`docs/ops/install-target-preflight.md`.

**Block 10 / Slice 10.11 (implementiert):** Begrenzter Payload-Install-Write via
`vyntrio-installer install` — erfordert `--force`, Preflight-Gates, Kopie der sechs
Allowlist-Payloads nur unter `distro/installer/target-sandbox/`; Post-Write-
Verifikation; kein Block-Device-Write, kein Service-Start. Vertrag:
`docs/ops/install-payload-write.md`.

**Block 10 / Slice 10.12 (implementiert):** Installer-Postflight und Operator-
Handover — strukturierte Zusammenfassung, `HANDOVER_RECORD.txt`, `vyntrio-installer
postflight`; ehrliche Deferred-Liste; kein Vollinstall-Anspruch. Vertrag:
`docs/ops/install-staged-workflow.md`.

**Block 10 / Slice 10.13 (implementiert):** Guarded Target-Mutation via
`vyntrio-installer install --apply-target` — erfordert `--force`, erfolgreiches
Preflight, explizite Disk-ID; mountet genau eine berechtigte, ungemountete
Partition (ext4/xfs/btrfs), kopiert dieselben sechs Payloads, Rollback bei Fehler;
`TARGET_MUTATION_RECORD.txt`; kein Partitionieren/Formatieren, kein Vollinstall.
Vertrag: `docs/ops/install-target-mutation.md`.

**Block 10 / Slice 10.14 (implementiert):** Dedizierter Existing-Partition-Apply-
Flow via `vyntrio-installer apply` — gleiche Gates wie `--apply-target`, eigenes
`installapply`-Orchestrierungspaket, `PARTITION_APPLY_RECORD.txt`, explizite
Mount/Copy/Rollback-Phasen; kein Sandbox-Fallback. Vertrag:
`docs/ops/install-partition-apply.md`.

**Block 10 / Slice 10.15 (implementiert):** Installer-Policy und Flow-
Konsolidierung — kanonisches Staged-Pipeline-Policy-Dokument, `vyntrio-installer
workflow`, gemeinsame CLI-Hilfetexte (`internal/platform/installpolicy`), klare
Grenzen zwischen verify/preflight/install/apply/postflight; explizit sekundär zum
USB-first-Auslieferungsweg. Vertrag: `docs/ops/installer-policy.md`.

## Primärer Produkt-Auslieferungsweg

| Phase | Block | Operator-Erlebnis |
|-------|-------|-------------------|
| 1 | 9 | USB Creator schreibt bootfähiges Install-Medium |
| 2 | 9 | Boot Linux-basiertes Vyntrio-Install-Image auf Zielhardware |
| 3 | 8 + 4 | Lokales Browser-Dashboard (`vyntrio-api`); Install-Assistent |
| 4 | 10 | Zielplatten-Installation aus Live-Session (interne Installer-Infra) |
| 5 | 4 | Bootstrap / Owner (ADR-0004) auf installiertem System |

**Fortschritt Block 9:** Payload-Staging + Envelope + Bootability-Foundation (9.11),
realer Boot-Chain (9.12), firmware-bootfähiges Raw-BIOS-Image (9.13), Runtime-Boot-
Harness + First-Boot-Pfad (9.14), Live-Rootfs-Userland mit chroot-verifiziertem
Dashboard (9.15), hardware-enabled Live-Initramfs (Storage/Netzwerk-Module +
DHCP, 9.16) und Image-Initrd-Swap auf das Live-Initramfs (9.17, sha256-bewiesen)
vorhanden — **Stage 1 abgeschlossen**. Stage 2 aktiv: Runtime-Boot-Verifikation
(Slice 10.1, read-only) implementiert, hier blockiert (kein qemu/`/dev/kvm`);
lokale Dashboard-Stabilität (Slice 10.2) und First-Boot-Onboarding-Klarheit
(Slice 10.3) — dashboard-first, busybox-sh First-Boot, beaufsichtigte WebGUI,
`/readyz`-Lokalnachweis, klare Boot→Dashboard→Login→Übersicht-Prompt in Produkt
und Landing-Page. LAN-Bind auf TLS blockiert. Offen für den Dashboard-on-Boot-
Nachweis: ein VM/Hardware-Boot-Harness; TLS vor LAN-Exposure; danach UEFI und
Host-USB-Creator.

**Nicht primär:** `distro/systemd/README.md` (manuell, Lab), Sandbox-`vyntrio-installer
install`, Make-`installer-*`-Skripte — Entwicklung und Übergang.

**Reihenfolge:** Block 9 bootfähiges Medium **vor** vollständiger Block-10-Produktreise;
bestehende Block-10-CLI ist wiederverwendbare Infrastruktur.

## Architekturregeln
- Keine Domain-Abhängigkeit auf Infrastrukturpakete.
- Jede externe Integration bekommt ein Interface und mindestens einen Adapter.
- Long-Running Operations laufen als Jobs mit Statusmodell.
- Breaking Changes benötigen ADR und Migrationskonzept.
