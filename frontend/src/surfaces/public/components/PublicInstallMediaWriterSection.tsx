import type { ReactNode } from "react";
import {
  formatInstallMediaBytes,
  type InstallMediaDto,
  type InstallMediaWriterArtifact,
} from "../../../features/public/installMediaDto";
import { formatSupportStatus } from "../../../features/public/installMediaFormatters";
import { useInstallMediaMetadata } from "../../../features/public/useInstallMediaMetadata";

type PublicInstallMediaWriterSectionProps = {
  heading: string;
  intro: string;
  honestNote: string;
  flowHeading: string;
  flowSteps: string[];
  commandHeading: string;
  listCommand: string;
  writeCommand: string;
  verifyHeading: string;
  verifyWindows: string;
  verifyMacOS: string;
  verifyLinux: string;
  buildNote: string;
  downloadsHeading: string;
  downloadsIntro: string;
  downloadUnavailable: string;
  guiDownloadsHeading?: string;
  cliDownloadsHeading?: string;
  headingId?: string;
};

const PLATFORM_LABELS: Record<string, string> = {
  windows: "Windows",
  macos: "macOS",
  linux: "Linux",
};

export function PublicInstallMediaWriterSection({
  heading,
  intro,
  honestNote,
  flowHeading,
  flowSteps,
  commandHeading,
  listCommand,
  writeCommand,
  verifyHeading,
  verifyWindows,
  verifyMacOS,
  verifyLinux,
  buildNote,
  downloadsHeading,
  downloadsIntro,
  downloadUnavailable,
  guiDownloadsHeading = "GUI media creator downloads",
  cliDownloadsHeading = "CLI writer downloads",
  headingId = "public-install-media-writer-heading",
}: PublicInstallMediaWriterSectionProps) {
  const metadataState = useInstallMediaMetadata();
  const metadata = metadataState.kind === "ready" ? metadataState.metadata : null;
  const writer = metadata?.writer;
  const artifact = metadata?.primary_artifact;
  const artifacts = writer?.artifacts?.length ? writer.artifacts : fallbackWriterArtifacts();
  const guiArtifacts = artifacts.filter(
    (item) => item.kind.startsWith("gui_") || item.kind.startsWith("native_")
  );
  const cliArtifacts = artifacts.filter(
    (item) => !item.kind.startsWith("gui_") && !item.kind.startsWith("native_")
  );

  return (
    <section className="vyn-public-media-writer" aria-labelledby={headingId}>
      <h2 id={headingId}>{heading}</h2>
      <p className="vyn-public-download-panel-intro">{intro}</p>
      <p className="vyn-public-media-creator-note" role="note">
        {honestNote}
      </p>

      <h3 className="vyn-public-download-subheading">Install media status</h3>
      <dl className="vyn-public-download-rows">
        <StatusRow label="Publication" value={metadata?.publication_status ?? "not_built"} />
        <StatusRow
          label="Support status"
          value={formatSupportStatus(metadata?.support_status)}
        />
        <StatusRow label="Release" value={formatRelease(metadata)} />
        <StatusRow label="Generated" value={metadata?.generated_at ?? "—"} />
        <StatusRow label="Primary artifact" value={artifact?.name ?? "vyntrio-install-media.img"} />
        <StatusRow
          label="Artifact type"
          value={
            artifact
              ? `${artifact.format} (${artifact.firmware_boot_mode})`
              : "raw_gpt_hybrid_disk (bios+uefi)"
          }
        />
        <StatusRow
          label="BIOS support"
          value={artifact?.bios_support || artifact?.firmware_boot_mode?.includes("bios") ? "yes" : "—"}
        />
        <StatusRow
          label="UEFI support"
          value={artifact?.uefi_support || artifact?.firmware_boot_mode?.includes("uefi") ? "yes" : "no — incomplete"}
        />
        <StatusRow
          label="Dual-mode"
          value={
            artifact?.dual_mode || artifact?.firmware_boot_mode === "bios+uefi"
              ? "yes (BIOS + UEFI)"
              : "no"
          }
        />
        <StatusRow
          label="File size"
          value={artifact?.size_bytes ? formatInstallMediaBytes(artifact.size_bytes) : "—"}
        />
        <StatusRow label="SHA-256" value={artifact?.sha256 ?? "—"} checksum />
        <StatusRow
          label="Image download"
          value={
            artifact?.download_available && artifact.download_path ? (
              <a href={artifact.download_path}>{artifact.name}</a>
            ) : (
              "Not staged on this host"
            )
          }
        />
        <StatusRow
          label="Media creator"
          value={writer?.name ?? "vyntrio-media-creator"}
        />
        <StatusRow
          label="Creator kind"
          value={
            writer
              ? `${writer.kind}${writer.gui_available ? " · GUI available" : ""}${
                  writer.native_gui ? " · native desktop (Tauri)" : " · not native desktop"
                }`
              : "native_desktop_tauri"
          }
        />
        <StatusRow
          label="Writer platforms"
          value={writer?.platforms?.join(", ") ?? "linux, windows"}
        />
      </dl>

      <h3 className="vyn-public-download-subheading">{downloadsHeading}</h3>
      <p className="vyn-public-download-panel-intro">{downloadsIntro}</p>

      <h4 className="vyn-public-download-subheading">{guiDownloadsHeading}</h4>
      <ArtifactList artifacts={guiArtifacts} downloadUnavailable={downloadUnavailable} />

      <h4 className="vyn-public-download-subheading">{cliDownloadsHeading}</h4>
      <ArtifactList artifacts={cliArtifacts} downloadUnavailable={downloadUnavailable} />

      {artifacts.some((item) => item.download_available && item.sha256) ? (
        <div className="vyn-public-media-writer-checksums">
          {artifacts
            .filter((item) => item.download_available && item.sha256)
            .map((item) => (
              <p key={item.name} className="vyn-public-download-panel-note">
                <code>{item.name}</code>{" "}
                <span className="vyn-public-media-creator-checksum">{item.sha256}</span>
              </p>
            ))}
        </div>
      ) : null}

      <h3 className="vyn-public-download-subheading">{flowHeading}</h3>
      <ol className="vyn-public-procedure-steps" aria-label={flowHeading}>
        {flowSteps.map((step, index) => (
          <li key={step} className="vyn-public-procedure-step">
            <span className="vyn-public-procedure-index">{String(index + 1)}</span>
            <div className="vyn-public-procedure-copy">
              <p>{step}</p>
            </div>
          </li>
        ))}
      </ol>

      <p className="vyn-public-media-creator-helper-label">{commandHeading}</p>
      <pre className="vyn-public-media-creator-command">{listCommand}</pre>
      <pre className="vyn-public-media-creator-command">{writeCommand}</pre>
      <p className="vyn-public-download-panel-note">{buildNote}</p>

      <h3 className="vyn-public-download-subheading">{verifyHeading}</h3>
      <p className="vyn-public-media-creator-helper-label">Windows (PowerShell)</p>
      <pre className="vyn-public-media-creator-command">{verifyWindows}</pre>
      <p className="vyn-public-media-creator-helper-label">macOS</p>
      <pre className="vyn-public-media-creator-command">{verifyMacOS}</pre>
      <p className="vyn-public-media-creator-helper-label">Linux</p>
      <pre className="vyn-public-media-creator-command">{verifyLinux}</pre>
    </section>
  );
}

function ArtifactList({
  artifacts,
  downloadUnavailable,
}: {
  artifacts: InstallMediaWriterArtifact[];
  downloadUnavailable: string;
}) {
  if (!artifacts.length) {
    return <p className="vyn-public-download-panel-note">{downloadUnavailable}</p>;
  }
  return (
    <dl className="vyn-public-download-rows">
      {artifacts.map((item) => (
        <div key={item.name} className="vyn-public-download-row">
          <dt>{`${PLATFORM_LABELS[item.platform] ?? item.platform} (${item.arch}) · ${item.kind}`}</dt>
          <dd>
            {item.download_available && item.download_path ? (
              <>
                <a href={item.download_path}>{item.name}</a>
                {item.size_bytes ? ` · ${formatInstallMediaBytes(item.size_bytes)}` : null}
              </>
            ) : (
              downloadUnavailable
            )}
          </dd>
        </div>
      ))}
    </dl>
  );
}

function fallbackWriterArtifacts(): InstallMediaWriterArtifact[] {
  return [
    {
      platform: "windows",
      arch: "amd64",
      name: "vyntrio-media-creator-windows-amd64-setup.exe",
      kind: "native_nsis_installer",
      download_available: false,
    },
    {
      platform: "linux",
      arch: "amd64",
      name: "vyntrio-media-creator-linux-amd64.deb",
      kind: "native_deb",
      download_available: false,
    },
    {
      platform: "linux",
      arch: "amd64",
      name: "vyntrio-media-creator-linux-amd64.AppImage",
      kind: "native_appimage",
      download_available: false,
    },
  ];
}

function formatRelease(metadata: InstallMediaDto | null | undefined): string {
  if (!metadata) {
    return "—";
  }
  return metadata.release.channel
    ? `${metadata.release.version} · ${metadata.release.channel}`
    : metadata.release.version;
}

function StatusRow({
  label,
  value,
  checksum = false,
}: {
  label: string;
  value: ReactNode;
  checksum?: boolean;
}) {
  return (
    <div className="vyn-public-download-row">
      <dt>{label}</dt>
      <dd className={checksum ? "vyn-public-media-creator-checksum" : undefined}>{value}</dd>
    </div>
  );
}
