import { Link } from "react-router-dom";
import {
  formatInstallMediaBytes,
  type InstallMediaDto,
} from "../../../features/public/installMediaDto";
import { formatSupportStatus } from "../../../features/public/installMediaFormatters";
import { useInstallMediaMetadata } from "../../../features/public/useInstallMediaMetadata";

export type PublicLiveReleaseBandProps = {
  eyebrow: string;
  heading: string;
  intro: string;
  downloadCta: string;
  downloadTo?: string;
  headingId?: string;
};

export function PublicLiveReleaseBand({
  eyebrow,
  heading,
  intro,
  downloadCta,
  downloadTo = "/download",
  headingId = "public-live-release-heading",
}: PublicLiveReleaseBandProps) {
  const state = useInstallMediaMetadata();
  const metadata = state.kind === "ready" ? state.metadata : null;
  const artifact = metadata?.primary_artifact;

  return (
    <section className="vyn-public-live-release" aria-labelledby={headingId}>
      <p className="vyn-public-section-eyebrow">{eyebrow}</p>
      <h2 id={headingId}>{heading}</h2>
      <p className="vyn-public-download-panel-intro">{intro}</p>
      <dl className="vyn-public-download-rows">
        <div className="vyn-public-download-row">
          <dt>Version</dt>
          <dd>{formatRelease(metadata)}</dd>
        </div>
        <div className="vyn-public-download-row">
          <dt>Support status</dt>
          <dd>{formatSupportStatus(metadata?.support_status)}</dd>
        </div>
        <div className="vyn-public-download-row">
          <dt>Install image</dt>
          <dd>{artifact?.name ?? "vyntrio-install-media.img"}</dd>
        </div>
        <div className="vyn-public-download-row">
          <dt>Type</dt>
          <dd>
            {artifact
              ? `${artifact.format} (${artifact.firmware_boot_mode})`
              : "raw_gpt_hybrid_disk (bios+uefi)"}
          </dd>
        </div>
        <div className="vyn-public-download-row">
          <dt>BIOS / UEFI / dual-mode</dt>
          <dd>
            {artifact?.dual_mode || artifact?.firmware_boot_mode === "bios+uefi"
              ? "BIOS yes · UEFI yes · dual-mode yes"
              : artifact
                ? `BIOS ${artifact.bios_support ? "yes" : "—"} · UEFI ${artifact.uefi_support ? "yes" : "no (incomplete)"} · dual-mode no`
                : "BIOS yes · UEFI yes · dual-mode yes (when staged)"}
          </dd>
        </div>
        <div className="vyn-public-download-row">
          <dt>Size</dt>
          <dd>{artifact?.size_bytes ? formatInstallMediaBytes(artifact.size_bytes) : "—"}</dd>
        </div>
        <div className="vyn-public-download-row">
          <dt>SHA-256</dt>
          <dd className="vyn-public-media-creator-checksum">{artifact?.sha256 ?? "—"}</dd>
        </div>
        <div className="vyn-public-download-row">
          <dt>Publication</dt>
          <dd>{metadata?.publication_status ?? "not_built"}</dd>
        </div>
      </dl>
      <p className="vyn-public-live-release-cta">
        <Link className="vyn-public-btn vyn-public-btn-primary" to={downloadTo}>
          {downloadCta}
        </Link>
      </p>
    </section>
  );
}

function formatRelease(metadata: InstallMediaDto | null | undefined): string {
  if (!metadata) {
    return "—";
  }
  return metadata.release.channel
    ? `${metadata.release.version} · ${metadata.release.channel}`
    : metadata.release.version;
}
