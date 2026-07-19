import {
  formatInstallMediaBytes,
  type InstallMediaWriterArtifact,
} from "../../../features/public/installMediaDto";
import { useInstallMediaMetadata } from "../../../features/public/useInstallMediaMetadata";

export type PublicPlatformCreatorHeroProps = {
  eyebrow: string;
  title: string;
  description: string;
  windowsCta: string;
  macosCta: string;
  linuxCta: string;
  macosBlockedNote: string;
  cardsHeading: string;
  cardsIntro: string;
  downloadUnavailable: string;
  titleId?: string;
};

function pickArtifact(
  artifacts: InstallMediaWriterArtifact[],
  platform: string,
  kindPrefix?: string
): InstallMediaWriterArtifact | undefined {
  return artifacts.find(
    (item) =>
      item.platform === platform &&
      item.download_available &&
      (!kindPrefix || item.kind.startsWith(kindPrefix) || item.kind.includes(kindPrefix))
  );
}

export function PublicPlatformCreatorHero({
  eyebrow,
  title,
  description,
  windowsCta,
  macosCta,
  linuxCta,
  macosBlockedNote,
  cardsHeading,
  cardsIntro,
  downloadUnavailable,
  titleId = "public-platform-creator-hero-title",
}: PublicPlatformCreatorHeroProps) {
  const state = useInstallMediaMetadata();
  const artifacts =
    state.kind === "ready" ? state.metadata.writer?.artifacts ?? [] : [];
  const windows = pickArtifact(artifacts, "windows", "native_nsis");
  const linuxDeb = pickArtifact(artifacts, "linux", "native_deb");
  const linuxApp = pickArtifact(artifacts, "linux", "native_appimage");
  const linux = linuxDeb ?? linuxApp;

  return (
    <section className="vyn-public-platform-hero" aria-labelledby={titleId}>
      <p className="vyn-public-section-eyebrow">{eyebrow}</p>
      <h1 id={titleId} className="vyn-public-platform-hero-title">
        {title}
      </h1>
      <p className="vyn-public-platform-hero-desc">{description}</p>

      <div className="vyn-public-platform-hero-ctas">
        {windows?.download_path ? (
          <a className="vyn-public-platform-cta" href={windows.download_path}>
            {windowsCta}
          </a>
        ) : (
          <span className="vyn-public-platform-cta is-disabled">{windowsCta}</span>
        )}
        <span className="vyn-public-platform-cta is-disabled" title={macosBlockedNote}>
          {macosCta}
        </span>
        {linux?.download_path ? (
          <a className="vyn-public-platform-cta" href={linux.download_path}>
            {linuxCta}
          </a>
        ) : (
          <span className="vyn-public-platform-cta is-disabled">{linuxCta}</span>
        )}
      </div>
      <p className="vyn-public-platform-hero-note">{macosBlockedNote}</p>

      <h2 className="vyn-public-download-subheading">{cardsHeading}</h2>
      <p className="vyn-public-download-panel-intro">{cardsIntro}</p>
      <div className="vyn-public-platform-cards">
        <PlatformCard
          platform="Windows"
          artifact={windows}
          unavailable={downloadUnavailable}
          hint="NSIS installer · native Tauri"
        />
        <PlatformCard
          platform="macOS"
          artifact={undefined}
          unavailable={macosBlockedNote}
          hint=".app / .dmg require a macOS build host"
        />
        <PlatformCard
          platform="Linux"
          artifact={linuxDeb}
          secondary={linuxApp}
          unavailable={downloadUnavailable}
          hint=".deb and AppImage · native Tauri"
        />
      </div>
    </section>
  );
}

function PlatformCard({
  platform,
  artifact,
  secondary,
  unavailable,
  hint,
}: {
  platform: string;
  artifact?: InstallMediaWriterArtifact;
  secondary?: InstallMediaWriterArtifact;
  unavailable: string;
  hint: string;
}) {
  return (
    <article className="vyn-public-platform-card">
      <h3>{platform}</h3>
      <p className="vyn-public-platform-card-hint">{hint}</p>
      {artifact?.download_available && artifact.download_path ? (
        <>
          <a className="vyn-public-btn vyn-public-btn-primary" href={artifact.download_path}>
            Download {artifact.name}
          </a>
          <p className="vyn-public-platform-card-meta">
            {formatInstallMediaBytes(artifact.size_bytes)}
            {artifact.sha256 ? ` · SHA-256 ${artifact.sha256.slice(0, 12)}…` : ""}
          </p>
        </>
      ) : (
        <p className="vyn-public-platform-card-meta">{unavailable}</p>
      )}
      {secondary?.download_available && secondary.download_path ? (
        <p>
          <a href={secondary.download_path}>{secondary.name}</a>
          {" · "}
          {formatInstallMediaBytes(secondary.size_bytes)}
        </p>
      ) : null}
    </article>
  );
}
