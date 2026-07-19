# TLS readiness for LAN dashboard bind (Stage 2 / Slice 10.4)

Operator and engineering contract for **Stage 2 / Slice 10.4** — removing the
TLS configuration blocker so the local dashboard can safely bind a non-loopback
(LAN) address without weakening the secure-by-default cookie policy.

## Purpose

Slice 10.2 recorded an honest blocker: non-loopback `bind_address` requires
`cookie_secure=true`, and the API previously served plain HTTP only. Slice 10.4
implements the minimum TLS path so LAN exposure is **possible** when explicitly
enabled, while keeping loopback HTTP as the default.

## Secure-by-default model

| Mode | How it is selected | Bind | Transport | Cookies |
|------|-------------------|------|-----------|---------|
| **Default (live media)** | Shipped `/etc/vyntrio/config.toml` | `127.0.0.1` | HTTP | `cookie_secure=false` |
| **LAN HTTPS (opt-in)** | Operator sets `VYNTRIO_LAN_BIND_IP` before/during first boot | Specific LAN IP | HTTPS (runtime self-signed cert) | `cookie_secure=true` |

Remote Connect, licensing, and a certificate manager remain **out of scope**.

## API / config changes

Optional TOML keys (paired):

- `tls_cert_file`
- `tls_key_file`

Validation (`internal/platform/config/tls.go`):

- Non-loopback `bind_address` → requires `cookie_secure=true` **and** both TLS paths.
- Loopback may use HTTP without TLS, or HTTPS with an optional cert pair.
- TLS files must exist and be readable at startup.

The HTTP server uses `ListenAndServeTLS` when both TLS paths are set
(`Config.TLSEnabled()`).

## Runtime wiring (live initramfs)

1. `scripts/prepare-live-dashboard-tls.sh` is staged into
   `live_root/usr/lib/vyntrio/` by `make install-media-runtime`.
2. `firstboot.sh` invokes it before starting `vyntrio-api`.
3. Without `VYNTRIO_LAN_BIND_IP`: records `tls_readiness: ready_loopback_only`;
   the shipped loopback config is used unchanged.
4. With `VYNTRIO_LAN_BIND_IP`: generates ephemeral cert/key under
   `/run/vyntrio/tls/` and writes `/run/vyntrio/dashboard-config.toml` for the
   supervised dashboard process.

`compose-live-rootfs.sh` optionally stages `openssl` into the live userland when
available on the build host.

## Commands

```bash
make build                          # rebuild vyntrio-api with TLS support
make install-media-runtime          # regenerate firstboot + TLS prep script
make install-media-live-rootfs      # compose userland; HTTP + HTTPS throwaway probes
make test-install-media-runtime     # asserts tls_readiness + no old TLS blocker
make test-install-media-live-rootfs # includes tls_readiness_test.sh
```

Build-time tunables: `VYNTRIO_LAN_BIND_IP` (runtime only), `VYNTRIO_DASHBOARD_PORT`
(default `8080`).

## Honest limits (not faked)

- **`dashboard_reachable: false` stays.** There is still no qemu/VM boot proof on
  this build host.
- **LAN HTTPS is not auto-enabled.** The operator must set `VYNTRIO_LAN_BIND_IP`.
- **Runtime certs are self-signed.** Suitable for local/LAN appliance use; not a
  public-PKI or Remote Connect story.
- Build-time HTTPS proof runs against a **throwaway chroot copy** with a probe
  cert; the shipped `live_root` stays state-free.

## What remains before Remote Connect

1. Booted-guest reachability proof (VM/hardware harness).
2. Owner bootstrap proven on the booted live runtime.
3. Trusted certificate / remote access architecture (Remote Connect is gated).

## References

- `scripts/prepare-live-dashboard-tls.sh`
- `scripts/verify-runtime-boot.sh` (`FIRST_BOOT.txt`, `RUNTIME_BOOT.txt`)
- `scripts/compose-live-rootfs.sh` (HTTPS `/` + `/readyz` probe)
- `internal/platform/config/tls.go`, `internal/platform/tlsutil/generate.go`
- `docs/ops/install-media-dashboard-stability.md` (Slice 10.2)
- `docs/03_ROADMAP.md`
