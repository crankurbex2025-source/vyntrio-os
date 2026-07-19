# Correction note — Windows media creator + website IA (2026-07-17)

## Windows failure cause (accepted)

1. PE subsystem was **CONSOLE (3)** — Explorer double-click opened a console, not an app.
2. GUI launch required argv0 to contain `media-creator` — renamed downloads printed usage and **exited 2** (reproduced under Wine with `/tmp/writer.exe`).
3. Browser open used `rundll32 url.dll,FileProtocolHandler` with errors ignored — Wine logged `ShellExecuteEx failed`.
4. macOS `.app.zip` was presented as a GUI app without a proven runnable path (Gatekeeper/unsigned).

## Correction applied

- Windows creator: `-H=windowsgui`, MessageBox + `cmd /c start`, fixed port `127.0.0.1:17823`, startup log, **no-args always launches GUI**.
- macOS `.app.zip` **withdrawn** (HTTP 404); Terminal helper binaries only.
- Website IA: nav Product/Download/Storage/Why/Docs; live release band; apps/VMs honesty; stronger product-truth.

## Still not claimed

- No physical Windows machine proof in this environment (PE/subsystem + launch-path fixes only).
- Not Electron/Qt native widgets; loopback web wizard.
- SmartScreen on unsigned Windows EXE may still warn — not treated as fixed by packaging alone.
