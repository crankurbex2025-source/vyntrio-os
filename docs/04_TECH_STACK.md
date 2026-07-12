# Tech Stack

## Betriebssystem
- Debian Stable
- Linux LTS Kernel
- systemd
- GRUB, initramfs

## Backend
- Go
- chi oder gin für HTTP Layer
- sqlc oder GORM-freier Query-Layer bevorzugt
- WebSocket Support
- Structured Logging

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
