# Vyntrio OS

## Produktüberblick
Vyntrio OS ist ein kommerzielles, Linux-basiertes Home-Server- und Private-Cloud-Betriebssystem für Privatanwender, Power User und kleine Unternehmen. Die Plattform kombiniert Storage, Container, Virtualisierung, Netzwerk, Backup, Monitoring, Automatisierung, App Marketplace und ein kommerzielles Lizenzmodell in einem einheitlichen Produkt.

## Produktziele
- Einfache Administration über ein modernes Web-Dashboard.
- Produktionsreife Architektur für x86_64 mit späterer ARM-Erweiterung.
- Modularer Unterbau in Go nach Clean-Architecture-Prinzipien.
- Sichere Update-, Lizenz- und Plugin-Mechanismen.
- Klare Trennung zwischen Community, Pro und Enterprise (Kapazitäts-/Geräte-Tiers — nicht fragmentierte Feature-Gates pro Modul).

## Primärer Auslieferungsweg (Produkt)

Der **primäre Installationsweg** für Endanwender ist:

1. **USB Creator** — Installationsmedium auf USB schreiben (Host-Tool, separat von `vyntrio-installer`).
2. **Bootfähiges Vyntrio-Install-Medium** — Linux-basiertes Live-/Install-Image (Block 9).
3. **Boot auf Zielhardware** — Firmware startet das Install-Medium.
4. **Lokales Browser-Dashboard** — Verwaltung und Install-Assistent im lokalen Web-UI (`vyntrio-api`).
5. **Zielplatten-Installation** — Ausführung **nach** Boot (Block 10 / interne Installer-Infrastruktur).
6. **First-Boot / Bootstrap** — Owner-Konto auf installiertem System (ADR-0004).

**Nicht primär:** manuelle systemd-Installation (`distro/systemd/README.md`) und Sandbox-/Entwickler-Pfade (`vyntrio-installer install` ohne Live-Boot) — nur Entwicklung, Lab und Übergang bis Block 9 bootfähig ist.

Siehe `docs/ADR/0006-appliance-install-recovery-media.md`, `docs/ops/installer-policy.md` (sekundär zum USB-Weg).

## Positionierung
Vyntrio OS positioniert sich zwischen Consumer-NAS-Lösungen, Self-Hosting-Plattformen und kleinen On-Prem-Infrastrukturlösungen. Das Produkt soll einfacher als klassische Linux-Server, flexibler als typische NAS-Oberflächen und kontrollierbarer als reine SaaS-Angebote sein.

## Leitprinzipien
1. Sicherheit vor Bequemlichkeit.
2. UX vor Linux-Komplexität.
3. Module statt Monolith ohne Grenzen.
4. Offline-fähige Kernfunktionen.
5. Kommerzielle Tragfähigkeit durch Geräte-/Server-Tiers und Ökosystem (nicht fragmentierte Per-Feature-Gates).

## Nicht-Ziele für Version 1.0
- Kein ARM-First-Release.
- Kein Kubernetes im Kernprodukt.
- Kein Multi-Tenant-Hyperscaler-Ansatz.
- Kein vollwertiger Groupware-Stack im Core.

## Erfolgskriterien
- Installierbare ISO mit stabilem First-Run-Setup.
- Funktionsfähige Storage-, Container- und VM-Verwaltung.
- Signierte Updates mit Rollback.
- Lizenzierung an Geräte-/Server-Tier gebunden (Kapazität/Edition), nicht an einzelne Feature-Schalter pro Modul.
- Dokumentierte APIs, Teststrategie und Release-Prozess.
