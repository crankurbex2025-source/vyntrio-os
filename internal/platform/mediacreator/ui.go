package mediacreator

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Vyntrio Media Creator</title>
  <link rel="stylesheet" href="/assets/app.css" />
</head>
<body>
  <main class="shell">
    <header class="brand">
      <p class="brand-mark">Vyntrio</p>
      <h1>Media Creator</h1>
      <p class="lede">Prepare bootable USB install media from the Vyntrio dual-mode (BIOS + UEFI) raw image.</p>
      <p id="meta" class="meta"></p>
    </header>

    <section class="panel" id="step-image">
      <h2>1. Install image</h2>
      <p class="hint">Select or enter the path to <code>vyntrio-install-media.img</code> on this computer (dual-mode GPT hybrid).</p>
      <label class="field">
        <span>Image path</span>
        <input id="image-path" type="text" spellcheck="false" autocomplete="off" placeholder="/path/to/vyntrio-install-media.img" />
      </label>
      <div class="row">
        <button type="button" id="btn-suggest" class="btn secondary">Find local .img</button>
        <button type="button" id="btn-load" class="btn primary">Load image</button>
      </div>
      <ul id="suggestions" class="suggestions" hidden></ul>
      <dl id="image-info" class="facts" hidden></dl>
    </section>

    <section class="panel" id="step-device">
      <h2>2. Target USB device</h2>
      <p class="hint">Only removable candidates are listed. Unmount the device before writing.</p>
      <div class="row">
        <button type="button" id="btn-refresh" class="btn secondary">Refresh devices</button>
      </div>
      <div id="devices" class="devices" role="listbox" aria-label="USB devices"></div>
    </section>

    <section class="panel warn" id="step-write">
      <h2>3. Write media</h2>
      <p class="warn-copy">Writing erases <strong>all data</strong> on the selected USB device.</p>
      <label class="check">
        <input id="confirm" type="checkbox" />
        <span>I understand this will overwrite the selected device.</span>
      </label>
      <div class="row">
        <button type="button" id="btn-dry" class="btn secondary">Dry run</button>
        <button type="button" id="btn-write" class="btn danger" disabled>Write image</button>
      </div>
      <div class="progress-wrap" hidden id="progress-wrap">
        <div class="progress-bar"><div id="progress-fill" class="progress-fill"></div></div>
        <p id="progress-label" class="progress-label">0%</p>
      </div>
      <p id="result" class="result" role="status"></p>
    </section>

    <section class="panel" id="step-next">
      <h2>4. Next boot</h2>
      <p id="boot-copy" class="hint">After a successful write, eject the USB safely, insert it into the target machine, and boot in UEFI or BIOS/legacy mode. Dual-mode media is the product baseline; BIOS-only images are incomplete.</p>
    </section>
  </main>
  <script src="/assets/app.js"></script>
</body>
</html>
`

const appCSS = `:root {
  --bg: #e8efe9;
  --bg-2: #d7e4db;
  --ink: #14201a;
  --muted: #4a5c52;
  --panel: rgba(255, 252, 247, 0.92);
  --line: #9bb0a2;
  --accent: #0f6b5c;
  --accent-2: #c45c26;
  --danger: #9b1d2f;
  --ok: #1f6b3a;
  --shadow: 0 18px 40px rgba(20, 32, 26, 0.12);
  --font-display: "Iowan Old Style", "Palatino Linotype", Palatino, "Book Antiqua", Georgia, serif;
  --font-body: "Segoe UI", "Helvetica Neue", Helvetica, Arial, sans-serif;
  --font-mono: "Cascadia Mono", "SF Mono", Consolas, monospace;
}
* { box-sizing: border-box; }
body {
  margin: 0;
  min-height: 100vh;
  color: var(--ink);
  font-family: var(--font-body);
  background:
    radial-gradient(1200px 600px at 10% -10%, #f4fff7 0%, transparent 55%),
    radial-gradient(900px 500px at 100% 0%, #ffe8d6 0%, transparent 50%),
    linear-gradient(160deg, var(--bg), var(--bg-2));
}
.shell {
  width: min(820px, calc(100% - 2rem));
  margin: 2rem auto 3rem;
  display: grid;
  gap: 1rem;
}
.brand-mark {
  margin: 0;
  font-family: var(--font-display);
  font-size: 2rem;
  letter-spacing: 0.04em;
  color: var(--accent);
}
.brand h1 {
  margin: 0.15rem 0 0.4rem;
  font-family: var(--font-display);
  font-weight: 600;
  font-size: 1.65rem;
}
.lede, .hint, .meta { color: var(--muted); line-height: 1.5; }
.meta { font-size: 0.9rem; }
.panel {
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 18px;
  padding: 1.15rem 1.25rem 1.3rem;
  box-shadow: var(--shadow);
}
.panel.warn { border-color: color-mix(in srgb, var(--accent-2) 55%, var(--line)); }
.panel h2 {
  margin: 0 0 0.55rem;
  font-family: var(--font-display);
  font-size: 1.2rem;
}
.field { display: grid; gap: 0.35rem; margin: 0.75rem 0; }
.field span { font-size: 0.85rem; color: var(--muted); }
input[type="text"] {
  width: 100%;
  border: 1px solid var(--line);
  border-radius: 10px;
  padding: 0.7rem 0.8rem;
  font: 0.95rem/1.3 var(--font-mono);
  background: #fff;
  color: var(--ink);
}
.row { display: flex; flex-wrap: wrap; gap: 0.6rem; margin-top: 0.6rem; }
.btn {
  border: 0;
  border-radius: 999px;
  padding: 0.65rem 1.05rem;
  font: 600 0.92rem/1 var(--font-body);
  cursor: pointer;
}
.btn:disabled { opacity: 0.45; cursor: not-allowed; }
.btn.primary { background: var(--accent); color: #f7fffb; }
.btn.secondary { background: #eef5f0; color: var(--ink); border: 1px solid var(--line); }
.btn.danger { background: var(--danger); color: #fff8f7; }
.suggestions { list-style: none; padding: 0; margin: 0.75rem 0 0; display: grid; gap: 0.4rem; }
.suggestions button {
  width: 100%;
  text-align: left;
  border: 1px dashed var(--line);
  border-radius: 10px;
  background: #f7fbf8;
  padding: 0.55rem 0.7rem;
  font: 0.85rem/1.3 var(--font-mono);
  cursor: pointer;
}
.facts { display: grid; grid-template-columns: 8rem 1fr; gap: 0.35rem 0.75rem; margin: 0.9rem 0 0; }
.facts dt { color: var(--muted); }
.facts dd { margin: 0; word-break: break-all; font-family: var(--font-mono); font-size: 0.85rem; }
.devices { display: grid; gap: 0.55rem; margin-top: 0.8rem; }
.device {
  display: grid;
  gap: 0.2rem;
  text-align: left;
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 0.75rem 0.85rem;
  background: #fbfefc;
  cursor: pointer;
  font: inherit;
  color: inherit;
}
.device[aria-selected="true"] {
  border-color: var(--accent);
  box-shadow: inset 0 0 0 1px var(--accent);
  background: #e8f6f1;
}
.device .name { font-weight: 650; }
.device .meta { font-size: 0.82rem; }
.check { display: flex; gap: 0.55rem; align-items: flex-start; margin: 0.8rem 0; }
.warn-copy { color: var(--accent-2); }
.progress-wrap { margin-top: 0.9rem; }
.progress-bar {
  height: 0.7rem;
  border-radius: 999px;
  background: #d9e6de;
  overflow: hidden;
}
.progress-fill {
  height: 100%;
  width: 0%;
  background: linear-gradient(90deg, var(--accent), #1f9b82);
  transition: width 120ms linear;
}
.progress-label, .result { margin: 0.45rem 0 0; font-size: 0.92rem; }
.result.ok { color: var(--ok); }
.result.err { color: var(--danger); }
code { font-family: var(--font-mono); font-size: 0.9em; }
`

const appJS = `(() => {
  const state = {
    imagePath: "",
    device: null,
    image: null,
  };

  const el = {
    meta: document.getElementById("meta"),
    imagePath: document.getElementById("image-path"),
    suggestions: document.getElementById("suggestions"),
    imageInfo: document.getElementById("image-info"),
    devices: document.getElementById("devices"),
    confirm: document.getElementById("confirm"),
    writeBtn: document.getElementById("btn-write"),
    progressWrap: document.getElementById("progress-wrap"),
    progressFill: document.getElementById("progress-fill"),
    progressLabel: document.getElementById("progress-label"),
    result: document.getElementById("result"),
    bootCopy: document.getElementById("boot-copy"),
  };

  function bytes(n) {
    if (!n && n !== 0) return "—";
    const units = ["B", "KB", "MB", "GB", "TB"];
    let v = Number(n);
    let i = 0;
    while (v >= 1024 && i < units.length - 1) { v /= 1024; i++; }
    return (i === 0 ? String(v) : v.toFixed(2)) + " " + units[i];
  }

  function setResult(text, ok) {
    el.result.textContent = text || "";
    el.result.className = "result" + (text ? (ok ? " ok" : " err") : "");
  }

  function syncWriteEnabled() {
    el.writeBtn.disabled = !(state.image && state.device && el.confirm.checked);
  }

  async function loadStatus() {
    const res = await fetch("/api/status");
    const data = await res.json();
    el.meta.textContent = [
      data.version,
      data.support_status,
      data.platform + "/" + data.arch,
      "local web GUI (not a signed native desktop framework app)",
    ].filter(Boolean).join(" · ");
    if (data.image_hint) el.imagePath.value = data.image_hint;
    if (data.boot_instruction) el.bootCopy.textContent = data.boot_instruction;
  }

  function renderImage(info) {
    state.image = info;
    state.imagePath = info.path;
    el.imageInfo.hidden = false;
    el.imageInfo.innerHTML =
      "<dt>Name</dt><dd>" + info.name + "</dd>" +
      "<dt>Size</dt><dd>" + bytes(info.size_bytes) + "</dd>" +
      "<dt>SHA-256</dt><dd>" + info.sha256 + "</dd>";
    syncWriteEnabled();
  }

  async function loadImage() {
    setResult("", true);
    const path = el.imagePath.value.trim();
    if (!path) {
      setResult("Enter an image path first.", false);
      return;
    }
    const res = await fetch("/api/image", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ path }),
    });
    const data = await res.json();
    if (!res.ok) {
      setResult(data.error || "Could not load image", false);
      return;
    }
    renderImage(data);
  }

  async function suggest() {
    const res = await fetch("/api/suggest");
    const data = await res.json();
    el.suggestions.innerHTML = "";
    if (!data.candidates || !data.candidates.length) {
      el.suggestions.hidden = false;
      el.suggestions.innerHTML = "<li><em>No vyntrio-install-media*.img found in cwd/Downloads/Desktop.</em></li>";
      return;
    }
    el.suggestions.hidden = false;
    data.candidates.forEach((path) => {
      const li = document.createElement("li");
      const btn = document.createElement("button");
      btn.type = "button";
      btn.textContent = path;
      btn.addEventListener("click", () => {
        el.imagePath.value = path;
        loadImage();
      });
      li.appendChild(btn);
      el.suggestions.appendChild(li);
    });
  }

  function renderDevices(devices) {
    el.devices.innerHTML = "";
    if (!devices.length) {
      el.devices.innerHTML = "<p class='hint'>No removable devices found. Insert a USB stick and refresh.</p>";
      return;
    }
    devices.forEach((device) => {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "device";
      btn.setAttribute("role", "option");
      btn.setAttribute("aria-selected", state.device && state.device.path === device.path ? "true" : "false");
      btn.innerHTML =
        "<span class='name'>" + device.name + "</span>" +
        "<span class='meta'>" + device.path + " · " + bytes(device.size_bytes) +
        (device.bus_type ? " · " + device.bus_type : "") +
        (device.mounted ? " · mounted" : "") + "</span>";
      btn.addEventListener("click", () => {
        state.device = device;
        renderDevices(devices);
        syncWriteEnabled();
      });
      el.devices.appendChild(btn);
    });
  }

  async function refreshDevices() {
    const res = await fetch("/api/devices");
    const data = await res.json();
    if (!res.ok) {
      setResult(data.error || "Device list failed", false);
      return;
    }
    renderDevices(data.devices || []);
  }

  async function runWrite(dryRun) {
    if (!state.image || !state.device) {
      setResult("Load an image and select a device first.", false);
      return;
    }
    if (!dryRun && !el.confirm.checked) {
      setResult("Confirm the destructive overwrite first.", false);
      return;
    }
    setResult(dryRun ? "Dry run in progress…" : "Writing…", true);
    el.progressWrap.hidden = false;
    el.progressFill.style.width = "0%";
    el.progressLabel.textContent = "0%";

    const res = await fetch("/api/write", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        image_path: state.image.path,
        device: state.device.path,
        confirm: !dryRun,
        dry_run: !!dryRun,
      }),
    });
    if (!res.ok || !res.body) {
      const data = await res.json().catch(() => ({}));
      setResult(data.error || "Write request failed", false);
      return;
    }

    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split("\n");
      buffer = lines.pop() || "";
      for (const line of lines) {
        if (!line.trim()) continue;
        let evt;
        try { evt = JSON.parse(line); } catch { continue; }
        if (evt.type === "progress" && evt.total_bytes) {
          const pct = Math.min(100, Math.round((evt.bytes_done / evt.total_bytes) * 100));
          el.progressFill.style.width = pct + "%";
          el.progressLabel.textContent = pct + "% · " + bytes(evt.bytes_done) + " / " + bytes(evt.total_bytes);
        }
        if (evt.type === "error") {
          setResult(evt.error || "Write failed", false);
        }
        if (evt.type === "complete") {
          el.progressFill.style.width = "100%";
          if (dryRun) {
            setResult("Dry run complete — no data written.", true);
          } else {
            setResult(
              "Write complete" + (evt.verified ? " and verified." : ".") +
              " " + (evt.boot_instruction || ""),
              true
            );
            if (evt.boot_instruction) el.bootCopy.textContent = evt.boot_instruction;
          }
        }
      }
    }
  }

  document.getElementById("btn-load").addEventListener("click", loadImage);
  document.getElementById("btn-suggest").addEventListener("click", suggest);
  document.getElementById("btn-refresh").addEventListener("click", refreshDevices);
  document.getElementById("btn-dry").addEventListener("click", () => runWrite(true));
  document.getElementById("btn-write").addEventListener("click", () => runWrite(false));
  el.confirm.addEventListener("change", syncWriteEnabled);

  loadStatus().then(refreshDevices).catch((err) => setResult(String(err), false));
})();
`
