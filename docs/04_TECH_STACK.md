# Tech Stack

## Betriebssystem
- Debian Stable
- Linux LTS Kernel
- systemd
- GRUB, initramfs

## Backend
- Go 1.24+
- **chi** (`go-chi/chi/v5`) für HTTP Layer — see `docs/ADR/0002-use-chi-router.md`
- **log/slog** (stdlib) für Structured Logging — see `docs/API_CONVENTIONS.md`
- sqlc oder GORM-freier Query-Layer bevorzugt (Slice 2+)
- WebSocket Support (later)

## Frontend
- React
- TypeScript
- TailwindCSS
- TanStack Query
- Zustand oder Redux Toolkit für UI-State
- Recharts oder ECharts für Dashboards

## Virtualisierung & Container
- Docker Engine + Compose
- KVM/QEMU/libvirt

## Storage & System
- EXT4, XFS, BTRFS
- smartmontools
- mdadm/LVM optional nach Scope
- rsync/restic für Backup-Integrationen

## Security
- Argon2id
- JWT/Refresh-Token-Modell
- TLS, nftables
- Paket- und Update-Signaturen via Ed25519

## Datenbank
- SQLite initial
- PostgreSQL später

## Qualität
- Go test, integration tests, Playwright, Vitest
- GitHub Actions
- golangci-lint, eslint, prettier
