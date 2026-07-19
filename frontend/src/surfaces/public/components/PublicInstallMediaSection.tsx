import {
  formatInstallMediaBytes,
  type InstallMediaDto,
} from "../../../features/public/installMediaDto";
import { formatSupportStatus } from "../../../features/public/installMediaFormatters";
import { useInstallMediaMetadata } from "../../../features/public/useInstallMediaMetadata";

type PublicInstallMediaSectionProps = {
  heading: string;
  intro: string;
  buildHeading: string;
  buildIntro: string;
  downloadHeading: string;
  limitationsHeading: string;
  statusLabels: {
    notBuilt: string;
    localStaging: string;
    unavailable: string;
    loading: string;
    error: string;
  };
  rowLabels: {
    installImage: string;
    imageFormat: string;
    checksum: string;
    verification: string;
    releaseChannel: string;
    buildTarget: string;
    stageTarget: string;
    download: string;
    generatedAt: string;
    supportStatus: string;
  };
  headingId?: string;
};

export function PublicInstallMediaSection({
  heading,
  intro,
  buildHeading,
  buildIntro,
  downloadHeading,
  limitationsHeading,
  statusLabels,
  rowLabels,
  headingId = "public-install-media-heading",
}: PublicInstallMediaSectionProps) {
  const state = useInstallMediaMetadata();

  const statusLabel =
    state.kind === "loading"
      ? statusLabels.loading
      : state.kind === "error"
        ? statusLabels.error
        : state.metadata.publication_status === "local_staging"
          ? statusLabels.localStaging
          : state.metadata.publication_status === "unavailable"
            ? statusLabels.unavailable
            : statusLabels.notBuilt;

  const rows =
    state.kind === "ready"
      ? buildRows(state.metadata, rowLabels, statusLabel)
      : buildFallbackRows(rowLabels, statusLabel);

  return (
    <section className="vyn-public-download-panel" aria-labelledby={headingId}>
      <h2 id={headingId}>{heading}</h2>
      <p className="vyn-public-download-panel-intro">{intro}</p>
      <dl className="vyn-public-download-rows">
        {rows.map((row) => (
          <div key={row.label} className="vyn-public-download-row">
            <dt>{row.label}</dt>
            <dd>{row.value}</dd>
          </div>
        ))}
      </dl>

      <h3 className="vyn-public-download-subheading">{buildHeading}</h3>
      <p className="vyn-public-download-panel-intro">{buildIntro}</p>
      {state.kind === "ready" ? (
        <>
          <p className="vyn-public-download-panel-note">
            <code>{state.metadata.build_target}</code> then{" "}
            <code>{state.metadata.stage_target}</code>
            {state.metadata.verify_command ? (
              <>
                {" "}
                · verify with <code>{state.metadata.verify_command}</code>
              </>
            ) : null}
          </p>
          {state.metadata.primary_artifact.download_available &&
          state.metadata.primary_artifact.download_path ? (
            <p className="vyn-public-download-panel-note">
              {downloadHeading}{" "}
              <a href={state.metadata.primary_artifact.download_path}>
                {state.metadata.primary_artifact.name}
              </a>
            </p>
          ) : null}
          {state.metadata.limitations.length > 0 ? (
            <>
              <h3 className="vyn-public-download-subheading">{limitationsHeading}</h3>
              <ul className="vyn-public-download-limitations">
                {state.metadata.limitations.map((item) => (
                  <li key={item}>{item}</li>
                ))}
              </ul>
            </>
          ) : null}
        </>
      ) : null}
    </section>
  );
}

function buildRows(
  metadata: InstallMediaDto,
  rowLabels: PublicInstallMediaSectionProps["rowLabels"],
  statusLabel: string
) {
  const artifact = metadata.primary_artifact;
  return [
    { label: rowLabels.installImage, value: artifact.name },
    { label: rowLabels.imageFormat, value: `${artifact.format} (${artifact.firmware_boot_mode})` },
    {
      label: "BIOS support",
      value: artifact.bios_support ? "yes" : artifact.firmware_boot_mode.includes("bios") ? "yes" : "no",
    },
    {
      label: "UEFI support",
      value: artifact.uefi_support ? "yes" : artifact.firmware_boot_mode.includes("uefi") ? "yes" : "no",
    },
    {
      label: "Dual-mode",
      value: artifact.dual_mode
        ? "yes (BIOS + UEFI)"
        : artifact.firmware_boot_mode === "bios+uefi"
          ? "yes (BIOS + UEFI)"
          : "no — incomplete if BIOS-only",
    },
    {
      label: "Secure Boot",
      value: artifact.secure_boot ?? "unsupported",
    },
    {
      label: "Media role",
      value: artifact.media_role ?? "appliance",
    },
    {
      label: "Image size",
      value: formatInstallMediaBytes(artifact.size_bytes),
    },
    {
      label: rowLabels.checksum,
      value: artifact.sha256 ?? "—",
    },
    {
      label: rowLabels.verification,
      value: metadata.verify_command ? "SHA-256 manifest (unsigned v1)" : statusLabel,
    },
    {
      label: rowLabels.releaseChannel,
      value: metadata.release.channel
        ? `${metadata.release.version} · ${metadata.release.channel}`
        : metadata.release.version,
    },
    {
      label: rowLabels.generatedAt,
      value: metadata.generated_at ?? statusLabel,
    },
    {
      label: rowLabels.supportStatus,
      value: formatSupportStatus(metadata.support_status),
    },
    { label: rowLabels.buildTarget, value: metadata.build_target },
    { label: rowLabels.stageTarget, value: metadata.stage_target },
    {
      label: rowLabels.download,
      value: artifact.download_available
        ? formatInstallMediaBytes(artifact.size_bytes)
        : statusLabel,
    },
  ];
}

function buildFallbackRows(
  rowLabels: PublicInstallMediaSectionProps["rowLabels"],
  statusLabel: string
) {
  return [
    { label: rowLabels.installImage, value: "vyntrio-install-media.img" },
    { label: rowLabels.imageFormat, value: "raw_gpt_hybrid_disk (bios+uefi)" },
    { label: "BIOS support", value: "yes (when staged)" },
    { label: "UEFI support", value: "required — yes when staged" },
    { label: "Dual-mode", value: "required product baseline" },
    { label: rowLabels.checksum, value: "—" },
    { label: rowLabels.verification, value: statusLabel },
    { label: rowLabels.releaseChannel, value: statusLabel },
    { label: rowLabels.generatedAt, value: statusLabel },
    { label: rowLabels.supportStatus, value: formatSupportStatus(undefined) },
    { label: rowLabels.buildTarget, value: "make install-media" },
    { label: rowLabels.stageTarget, value: "make release-install-media-stage" },
    { label: rowLabels.download, value: statusLabel },
  ];
}
