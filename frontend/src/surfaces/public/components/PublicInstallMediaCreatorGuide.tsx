import { formatInstallMediaBytes } from "../../../features/public/installMediaDto";
import { useInstallMediaMetadata } from "../../../features/public/useInstallMediaMetadata";

export type PublicInstallMediaCreatorStep = {
  step: string;
  title: string;
  body: string;
};

export type PublicInstallMediaCreatorGuideProps = {
  heading: string;
  intro: string;
  earlyAccessNote: string;
  artifactHeading: string;
  warningsHeading: string;
  warnings: string[];
  usbHeading: string;
  usbIntro: string;
  usbSteps: PublicInstallMediaCreatorStep[];
  usbHelperLabel: string;
  usbHelperCommand: string;
  vmHeading: string;
  vmIntro: string;
  vmSteps: PublicInstallMediaCreatorStep[];
  vmHelperLabel: string;
  vmHelperCommand: string;
  afterBootHeading: string;
  afterBootBody: string;
  checksumPending: string;
  downloadPending: string;
  headingId?: string;
};

export function PublicInstallMediaCreatorGuide({
  heading,
  intro,
  earlyAccessNote,
  artifactHeading,
  warningsHeading,
  warnings,
  usbHeading,
  usbIntro,
  usbSteps,
  usbHelperLabel,
  usbHelperCommand,
  vmHeading,
  vmIntro,
  vmSteps,
  vmHelperLabel,
  vmHelperCommand,
  afterBootHeading,
  afterBootBody,
  checksumPending,
  downloadPending,
  headingId = "public-install-media-creator-heading",
}: PublicInstallMediaCreatorGuideProps) {
  const metadataState = useInstallMediaMetadata();
  const artifact =
    metadataState.kind === "ready" ? metadataState.metadata.primary_artifact : null;
  const checksum = artifact?.sha256 ?? checksumPending;
  const downloadAvailable = artifact?.download_available === true && artifact.download_path;
  const sizeLabel = artifact?.size_bytes
    ? formatInstallMediaBytes(artifact.size_bytes)
    : downloadPending;

  return (
    <section className="vyn-public-media-creator" aria-labelledby={headingId}>
      <h2 id={headingId}>{heading}</h2>
      <p className="vyn-public-download-panel-intro">{intro}</p>
      <p className="vyn-public-media-creator-note" role="note">
        {earlyAccessNote}
      </p>

      <div className="vyn-public-media-creator-artifact">
        <h3 className="vyn-public-download-subheading">{artifactHeading}</h3>
        <dl className="vyn-public-download-rows">
          <div className="vyn-public-download-row">
            <dt>Image</dt>
            <dd>{artifact?.name ?? "vyntrio-install-media.img"}</dd>
          </div>
          <div className="vyn-public-download-row">
            <dt>Size</dt>
            <dd>{sizeLabel}</dd>
          </div>
          <div className="vyn-public-download-row">
            <dt>SHA-256</dt>
            <dd className="vyn-public-media-creator-checksum">{checksum}</dd>
          </div>
          <div className="vyn-public-download-row">
            <dt>Download</dt>
            <dd>
              {downloadAvailable ? (
                <a href={artifact!.download_path}>{artifact!.name}</a>
              ) : (
                downloadPending
              )}
            </dd>
          </div>
        </dl>
      </div>

      <h3 className="vyn-public-download-subheading">{warningsHeading}</h3>
      <ul className="vyn-public-download-limitations">
        {warnings.map((warning) => (
          <li key={warning}>{warning}</li>
        ))}
      </ul>

      <div className="vyn-public-media-creator-path">
        <h3 className="vyn-public-download-subheading">{usbHeading}</h3>
        <p className="vyn-public-download-panel-intro">{usbIntro}</p>
        <CreatorSteps steps={usbSteps} ariaLabel={usbHeading} />
        <p className="vyn-public-media-creator-helper-label">{usbHelperLabel}</p>
        <pre className="vyn-public-media-creator-command">{usbHelperCommand}</pre>
      </div>

      <div className="vyn-public-media-creator-path">
        <h3 className="vyn-public-download-subheading">{vmHeading}</h3>
        <p className="vyn-public-download-panel-intro">{vmIntro}</p>
        <CreatorSteps steps={vmSteps} ariaLabel={vmHeading} />
        <p className="vyn-public-media-creator-helper-label">{vmHelperLabel}</p>
        <pre className="vyn-public-media-creator-command">{vmHelperCommand}</pre>
      </div>

      <div className="vyn-public-media-creator-after-boot">
        <h3 className="vyn-public-download-subheading">{afterBootHeading}</h3>
        <p className="vyn-public-download-panel-intro">{afterBootBody}</p>
      </div>
    </section>
  );
}

function CreatorSteps({
  steps,
  ariaLabel,
}: {
  steps: PublicInstallMediaCreatorStep[];
  ariaLabel: string;
}) {
  return (
    <ol className="vyn-public-procedure-steps" aria-label={ariaLabel}>
      {steps.map((step) => (
        <li key={step.step} className="vyn-public-procedure-step">
          <span className="vyn-public-procedure-index">{step.step}</span>
          <div className="vyn-public-procedure-copy">
            <h4>{step.title}</h4>
            <p>{step.body}</p>
          </div>
        </li>
      ))}
    </ol>
  );
}
