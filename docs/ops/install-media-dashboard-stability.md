# Local dashboard stability (Stage 2 / Slice 10.2)

Operator and engineering contract for **Stage 2 / Slice 10.2** — stabilizing the
path from a booted image to the **local browser dashboard (WebGUI)**, so Vyntrio
behaves like a local-first appliance where the WebGUI is the primary management
surface (Unraid-style local WebGUI first, remote later).

## Purpose

Stage 1 made the image boot the Vyntrio live initramfs. Slice 10.2 makes the
**dashboard itself** behave like an appliance service on that live runtime:

1. **The first-boot entrypoint actually runs.** `firstboot.sh` previously used
   `#!/usr/bin/env bash`, but the live initramfs ships **busybox only** (no bash,
   no `env` applet), so the entrypoint could not exec and the dashboard never
   started on boot. It now runs under `/bin/sh` (busybox) using POSIX syntax.
2. **The dashboard is supervised.** `firstboot.sh` runs `vyntrio-api` in a bounded
   respawn loop (default 5 restarts) so the WebGUI stays up; a hard-failing binary
   drops to a diagnostic shell instead of disappearing or spinning forever.
3. **Local readiness is proven at build time.** The live-rootfs throwaway probe
   now checks `/readyz` (backend/DB ready), not just `/` (static UI) — stronger
   evidence that the dashboard reaches a stable, serving state.

## Commands

```bash
make install-media-runtime        # regenerates the hardened firstboot.sh + records
make test-install-media-runtime   # asserts busybox-sh, supervision, honest LAN/TLS blocker
make install-media-live-rootfs    # composes userland; probes / and /readyz (throwaway copy)
make test-install-media-live-rootfs
```

Tunables: `VYNTRIO_DASHBOARD_PORT` (default `8080`),
`VYNTRIO_DASHBOARD_MAX_RESTARTS` (default `5`).

## Dashboard-first model

- The local WebGUI (`vyntrio-api`, embedded SPA) is the **primary** management
  interface — not a side feature. `RUNTIME_BOOT.txt` records
  `dashboard_primary_surface: true` and `dashboard_supervised: true`.

## Honest limits (not faked)

- **Loopback HTTP is the default bind.** LAN HTTPS is opt-in via
  `VYNTRIO_LAN_BIND_IP` and runtime TLS prep (Slice 10.4). See
  `docs/ops/install-media-dashboard-tls.md`.
- **Boot-to-dashboard reachability is not proven here.** There is no `qemu`/`/dev/kvm`
  on this build host, so a real booted-guest HTTP proof cannot run (see
  `docs/ops/install-media-runtime-verify.md`). `dashboard_reachable: false` stays.
- The build-time `/readyz` proof runs against a **throwaway chroot copy**; the
  shipped `live_root` stays state-free (no DB, credentials, or secrets).

## What Slice 10.4 added (TLS / LAN bind readiness)

- `vyntrio-api` accepts optional `tls_cert_file` / `tls_key_file` and serves
  HTTPS when configured; non-loopback bind requires `cookie_secure=true` + TLS.
- `prepare-live-dashboard-tls.sh` generates ephemeral runtime certs under
  `/run/vyntrio` when the operator sets `VYNTRIO_LAN_BIND_IP`.
- Provenance records now use `tls_readiness: ready` instead of the prior
  `lan_exposure_blocked_on: tls_required_for_non_loopback_bind` config blocker.

## What remains missing before Remote Connect

1. A **VM/hardware boot harness** to prove the supervised dashboard is reachable
   after a real boot (Slice 10.1 records exact blockers).
2. First-boot **owner bootstrap** proven on the booted live runtime.
3. Remote Connect architecture (trusted certs, remote path) — **gated**, not started.

## References

- `scripts/verify-runtime-boot.sh` (generates `firstboot.sh`, `FIRST_BOOT.txt`, `RUNTIME_BOOT.txt`)
- `scripts/compose-live-rootfs.sh` (`/` + `/readyz` local probe)
- `distro/install-media/bootability-contract.md`
- `docs/ops/install-media-dashboard-tls.md` (Slice 10.4)
- `docs/03_ROADMAP.md` (single active development line)
