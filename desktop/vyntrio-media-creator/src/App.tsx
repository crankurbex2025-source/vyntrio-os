import { useCallback, useEffect, useMemo, useState } from "react";
import { invoke } from "@tauri-apps/api/core";
import { listen } from "@tauri-apps/api/event";
import "./App.css";

type Locale = "en" | "de";
type Step = "welcome" | "release" | "storage" | "confirm" | "progress" | "done" | "help";

type ReleaseArtifact = {
  name: string;
  format: string;
  firmware_boot_mode: string;
  bios_support: boolean;
  uefi_support: boolean;
  dual_mode: boolean;
  secure_boot?: string;
  media_role?: string;
  size_bytes?: number;
  sha256?: string;
  download_available: boolean;
  download_path?: string;
  version: string;
  build_id?: string;
  channel?: string;
  support_status?: string;
  generated_at?: string;
  source: string;
  latest?: boolean;
};

type StorageDevice = {
  id: string;
  path: string;
  name: string;
  size_bytes: number;
  removable: boolean;
  bus_type: string;
  mounted: boolean;
  mount_points: string[];
};

type WriteProgress = {
  bytes_written: number;
  total_bytes: number;
  percent: number;
  phase: string;
};

type WriteResult = {
  device_path: string;
  bytes_written: number;
  verified: boolean;
  image_sha256: string;
  message: string;
};

const copy = {
  en: {
    brand: "Vyntrio",
    title: "Media Creator",
    help: "Help",
    info: "About",
    language: "Language",
    welcomeTitle: "Prepare bootable appliance media",
    welcomeBody:
      "Write the Vyntrio USB appliance image to a flash drive. The USB boots the real Vyntrio system (UEFI required; BIOS fallback when available). Secure Boot is unsupported.",
    start: "Get started",
    releaseTitle: "Select image version",
    releaseHint: "Versions come from Vyntrio release metadata — latest stable plus older published builds when available.",
    bios: "BIOS",
    uefi: "UEFI",
    dual: "Dual-mode",
    incomplete: "Incomplete — BIOS-only is not the product baseline",
    storageTitle: "Select USB storage",
    storageHint: "Only removable candidates are listed. Unmount before writing.",
    refresh: "Refresh devices",
    mounted: "Mounted — unmount first",
    confirmTitle: "Confirm destructive write",
    confirmBody: "This overwrites the selected USB device completely. All existing data will be destroyed.",
    imagePath: "Local image path",
    findImages: "Find local .img",
    write: "Write media",
    back: "Back",
    next: "Next",
    progressTitle: "Writing install media",
    doneTitle: "Complete",
    failureTitle: "Write failed",
    bootNext:
      "Eject the USB safely, then boot the target machine in UEFI (preferred) or BIOS/legacy mode.",
    helpTitle: "Help & about",
    helpBody:
      "Vyntrio Media Creator is a native desktop app (Tauri). It is not a browser-only wizard. USB writes require elevation. Secure Boot is not signed on engineering media.",
    noDevices: "No removable USB devices found.",
    noReleases: "No release metadata available.",
    size: "Size",
    version: "Version",
    checksum: "SHA-256",
    support: "Support status",
    verified: "Prefix verification",
    yes: "yes",
    no: "no",
  },
  de: {
    brand: "Vyntrio",
    title: "Media Creator",
    help: "Hilfe",
    info: "Info",
    language: "Sprache",
    welcomeTitle: "Bootfähiges Appliance-Medium vorbereiten",
    welcomeBody:
      "Schreibe das Vyntrio-USB-Appliance-Image auf einen Stick. Der Stick bootet das echte Vyntrio-System (UEFI erforderlich; BIOS-Fallback wenn verfügbar). Secure Boot wird nicht unterstützt.",
    start: "Starten",
    releaseTitle: "Image-Version wählen",
    releaseHint:
      "Versionen kommen aus den Vyntrio-Release-Metadaten — neueste plus ältere veröffentlichte Builds, wenn vorhanden.",
    bios: "BIOS",
    uefi: "UEFI",
    dual: "Dual-Mode",
    incomplete: "Unvollständig — BIOS-only ist nicht die Produktbasis",
    storageTitle: "USB-Speicher wählen",
    storageHint: "Nur Wechseldatenträger. Vor dem Schreiben aushängen.",
    refresh: "Geräte aktualisieren",
    mounted: "Eingehängt — zuerst aushängen",
    confirmTitle: "Destruktives Schreiben bestätigen",
    confirmBody: "Das gewählte USB-Gerät wird vollständig überschrieben. Alle Daten gehen verloren.",
    imagePath: "Lokaler Image-Pfad",
    findImages: "Lokale .img suchen",
    write: "Medium schreiben",
    back: "Zurück",
    next: "Weiter",
    progressTitle: "Installationsmedium wird geschrieben",
    doneTitle: "Fertig",
    failureTitle: "Schreiben fehlgeschlagen",
    bootNext:
      "USB sicher entfernen und Zielrechner im UEFI- (bevorzugt) oder BIOS/Legacy-Modus booten.",
    helpTitle: "Hilfe & Info",
    helpBody:
      "Vyntrio Media Creator ist eine native Desktop-App (Tauri), kein Browser-Wizard. USB-Schreiben braucht erhöhte Rechte. Secure Boot ist auf Engineering-Media unsigniert.",
    noDevices: "Keine Wechseldatenträger gefunden.",
    noReleases: "Keine Release-Metadaten verfügbar.",
    size: "Größe",
    version: "Version",
    checksum: "SHA-256",
    support: "Support-Status",
    verified: "Präfix-Verifikation",
    yes: "ja",
    no: "nein",
  },
} as const;

function formatBytes(n?: number) {
  if (!n) return "—";
  if (n < 1024) return `${n} B`;
  if (n < 1024 ** 2) return `${(n / 1024).toFixed(1)} KiB`;
  if (n < 1024 ** 3) return `${(n / 1024 ** 2).toFixed(1)} MiB`;
  return `${(n / 1024 ** 3).toFixed(2)} GiB`;
}

export default function App() {
  const [locale, setLocale] = useState<Locale>("en");
  const t = copy[locale];
  const [step, setStep] = useState<Step>("welcome");
  const [releases, setReleases] = useState<ReleaseArtifact[]>([]);
  const [selectedRelease, setSelectedRelease] = useState<ReleaseArtifact | null>(null);
  const [devices, setDevices] = useState<StorageDevice[]>([]);
  const [selectedDevice, setSelectedDevice] = useState<StorageDevice | null>(null);
  const [imagePath, setImagePath] = useState("");
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [progress, setProgress] = useState<WriteProgress | null>(null);
  const [result, setResult] = useState<WriteResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [appVersion, setAppVersion] = useState("0.2.0");

  const loadReleases = useCallback(async () => {
    try {
      const list = await invoke<ReleaseArtifact[]>("list_install_releases", { metadataUrl: null });
      setReleases(list);
      if (list[0]) setSelectedRelease(list[0]);
    } catch (e) {
      setError(String(e));
    }
  }, []);

  const loadDevices = useCallback(async () => {
    try {
      const list = await invoke<StorageDevice[]>("list_storage_devices");
      setDevices(list);
    } catch (e) {
      setError(String(e));
    }
  }, []);

  useEffect(() => {
    invoke<{ version?: string }>("get_app_info")
      .then((info) => setAppVersion(info.version ?? "0.2.0"))
      .catch(() => undefined);
    loadReleases();
    const unlisten = listen<WriteProgress>("write-progress", (event) => {
      setProgress(event.payload);
    });
    return () => {
      unlisten.then((fn) => fn());
    };
  }, [loadReleases]);

  const firmwareLabel = useMemo(() => {
    if (!selectedRelease) return "";
    if (selectedRelease.dual_mode && selectedRelease.uefi_support) {
      return `${t.bios}: ${t.yes} · ${t.uefi}: ${t.yes} · ${t.dual}: ${t.yes}`;
    }
    if (!selectedRelease.uefi_support) return t.incomplete;
    return `${t.bios}: ${selectedRelease.bios_support ? t.yes : t.no} · ${t.uefi}: ${
      selectedRelease.uefi_support ? t.yes : t.no
    }`;
  }, [selectedRelease, t]);

  async function onFindImages() {
    const list = await invoke<string[]>("suggest_images");
    setSuggestions(list);
    if (list[0] && !imagePath) setImagePath(list[0]);
  }

  async function onWrite() {
    if (!selectedDevice || !imagePath) return;
    setError(null);
    setResult(null);
    setProgress({ bytes_written: 0, total_bytes: 0, percent: 0, phase: "starting" });
    setStep("progress");
    try {
      const res = await invoke<WriteResult>("write_install_media", {
        imagePath,
        devicePath: selectedDevice.path,
        expectedSha256: selectedRelease?.sha256 ?? null,
      });
      setResult(res);
      setStep("done");
    } catch (e) {
      setError(String(e));
      setStep("done");
    }
  }

  return (
    <div className="shell">
      <header className="topbar">
        <div className="brand-block">
          <p className="brand">{t.brand}</p>
          <h1>{t.title}</h1>
        </div>
        <div className="top-actions">
          <button type="button" className="linkish" onClick={() => setStep("help")}>
            {t.help}
          </button>
          <label className="lang">
            <span>{t.language}</span>
            <select value={locale} onChange={(e) => setLocale(e.target.value as Locale)}>
              <option value="en">EN</option>
              <option value="de">DE</option>
            </select>
          </label>
        </div>
      </header>
      <div className="accent-line" />

      <main className="stage">
        {step === "welcome" && (
          <section className="panel">
            <h2>{t.welcomeTitle}</h2>
            <p>{t.welcomeBody}</p>
            <p className="meta">v{appVersion} · native Tauri desktop</p>
            <div className="actions">
              <button type="button" className="primary" onClick={() => setStep("release")}>
                {t.start}
              </button>
            </div>
          </section>
        )}

        {step === "release" && (
          <section className="panel">
            <h2>{t.releaseTitle}</h2>
            <p className="hint">{t.releaseHint}</p>
            {releases.length === 0 ? (
              <p>{t.noReleases}</p>
            ) : (
              <ul className="choice-list">
                {releases.map((r) => (
                  <li key={`${r.source}-${r.name}-${r.sha256 ?? "none"}`}>
                    <button
                      type="button"
                      className={selectedRelease === r ? "choice selected" : "choice"}
                      onClick={() => setSelectedRelease(r)}
                    >
                      <strong>
                        {r.latest ? "Latest · " : ""}
                        {r.version}
                        {r.build_id ? ` (${r.build_id})` : ""}
                        {r.channel ? ` · ${r.channel}` : ""}
                      </strong>
                      <span>{r.name}</span>
                      <span>
                        {r.format} ({r.firmware_boot_mode})
                        {r.media_role ? ` · ${r.media_role}` : ""}
                      </span>
                      <span>
                        {t.size}: {formatBytes(r.size_bytes)} · SHA-256:{" "}
                        {r.sha256 ? `${r.sha256.slice(0, 12)}…` : "—"}
                      </span>
                      <span>
                        UEFI: {r.uefi_support ? t.yes : t.no} · BIOS:{" "}
                        {r.bios_support ? t.yes : t.no} · Secure Boot:{" "}
                        {r.secure_boot ?? "unsupported"}
                      </span>
                      <span>
                        {t.support}: {r.support_status ?? "—"}
                        {r.generated_at ? ` · ${r.generated_at}` : ""}
                      </span>
                      <span className={r.uefi_support && r.dual_mode ? "ok" : "warn"}>
                        {r.uefi_support && r.dual_mode
                          ? `${t.bios}/${t.uefi}/${t.dual}: ${t.yes}`
                          : t.incomplete}
                      </span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
            <div className="actions">
              <button type="button" className="ghost" onClick={() => setStep("welcome")}>
                {t.back}
              </button>
              <button
                type="button"
                className="primary"
                disabled={!selectedRelease || !selectedRelease.uefi_support}
                onClick={async () => {
                  await loadDevices();
                  setStep("storage");
                }}
              >
                {t.next}
              </button>
            </div>
          </section>
        )}

        {step === "storage" && (
          <section className="panel">
            <h2>{t.storageTitle}</h2>
            <p className="hint">{t.storageHint}</p>
            <div className="actions inline">
              <button type="button" className="ghost" onClick={loadDevices}>
                {t.refresh}
              </button>
              <button type="button" className="ghost" onClick={onFindImages}>
                {t.findImages}
              </button>
            </div>
            <label className="field">
              <span>{t.imagePath}</span>
              <input
                value={imagePath}
                onChange={(e) => setImagePath(e.target.value)}
                placeholder="/path/to/vyntrio-install-media.img"
              />
            </label>
            {suggestions.length > 0 && (
              <ul className="suggestions">
                {suggestions.map((s) => (
                  <li key={s}>
                    <button type="button" className="linkish" onClick={() => setImagePath(s)}>
                      {s}
                    </button>
                  </li>
                ))}
              </ul>
            )}
            {devices.length === 0 ? (
              <p>{t.noDevices}</p>
            ) : (
              <ul className="choice-list">
                {devices.map((d) => (
                  <li key={d.id}>
                    <button
                      type="button"
                      className={selectedDevice?.id === d.id ? "choice selected" : "choice"}
                      onClick={() => setSelectedDevice(d)}
                      disabled={d.mounted}
                    >
                      <strong>{d.name}</strong>
                      <span>
                        {d.path} · {formatBytes(d.size_bytes)} · {d.bus_type}
                      </span>
                      {d.mounted && <span className="warn">{t.mounted}</span>}
                    </button>
                  </li>
                ))}
              </ul>
            )}
            <div className="actions">
              <button type="button" className="ghost" onClick={() => setStep("release")}>
                {t.back}
              </button>
              <button
                type="button"
                className="primary"
                disabled={!selectedDevice || !imagePath || selectedDevice.mounted}
                onClick={() => setStep("confirm")}
              >
                {t.next}
              </button>
            </div>
          </section>
        )}

        {step === "confirm" && selectedRelease && selectedDevice && (
          <section className="panel">
            <h2>{t.confirmTitle}</h2>
            <p className="warn-box">{t.confirmBody}</p>
            <dl className="facts">
              <div>
                <dt>{t.version}</dt>
                <dd>{selectedRelease.version}</dd>
              </div>
              <div>
                <dt>Image</dt>
                <dd>{imagePath}</dd>
              </div>
              <div>
                <dt>USB</dt>
                <dd>
                  {selectedDevice.name} ({selectedDevice.path})
                </dd>
              </div>
              <div>
                <dt>Firmware</dt>
                <dd>{firmwareLabel}</dd>
              </div>
              <div>
                <dt>{t.checksum}</dt>
                <dd className="mono">{selectedRelease.sha256 ?? "—"}</dd>
              </div>
            </dl>
            <div className="actions">
              <button type="button" className="ghost" onClick={() => setStep("storage")}>
                {t.back}
              </button>
              <button type="button" className="danger" onClick={onWrite}>
                {t.write}
              </button>
            </div>
          </section>
        )}

        {step === "progress" && (
          <section className="panel">
            <h2>{t.progressTitle}</h2>
            <p className="hint">{progress?.phase ?? "…"}</p>
            <div className="progress-track">
              <div className="progress-fill" style={{ width: `${progress?.percent ?? 0}%` }} />
            </div>
            <p>
              {formatBytes(progress?.bytes_written)} / {formatBytes(progress?.total_bytes)} (
              {(progress?.percent ?? 0).toFixed(1)}%)
            </p>
          </section>
        )}

        {step === "done" && (
          <section className="panel">
            <h2>{error ? t.failureTitle : t.doneTitle}</h2>
            {error ? <p className="warn-box">{error}</p> : <p className="ok-box">{result?.message}</p>}
            {result && (
              <dl className="facts">
                <div>
                  <dt>{t.verified}</dt>
                  <dd>{result.verified ? t.yes : t.no}</dd>
                </div>
                <div>
                  <dt>{t.checksum}</dt>
                  <dd className="mono">{result.image_sha256}</dd>
                </div>
              </dl>
            )}
            {!error && <p>{t.bootNext}</p>}
            <div className="actions">
              <button type="button" className="primary" onClick={() => setStep("welcome")}>
                {t.start}
              </button>
            </div>
          </section>
        )}

        {step === "help" && (
          <section className="panel">
            <h2>{t.helpTitle}</h2>
            <p>{t.helpBody}</p>
            <p className="meta">
              {t.brand} Media Creator v{appVersion}
            </p>
            <div className="actions">
              <button type="button" className="primary" onClick={() => setStep("welcome")}>
                {t.back}
              </button>
            </div>
          </section>
        )}
      </main>

      <footer className="footer">
        <button type="button" className="linkish" onClick={() => setStep("help")}>
          {t.info}
        </button>
        <span className="meta">UEFI dual-mode baseline · engineering early access</span>
      </footer>
    </div>
  );
}
