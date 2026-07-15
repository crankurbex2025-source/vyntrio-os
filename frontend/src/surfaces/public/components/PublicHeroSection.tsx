import { Link } from "react-router-dom";

export type PublicHeroSectionProps = {
  eyebrow: string;
  title: string;
  description: string;
  ctaDownloadLabel?: string;
  ctaDownloadHint?: string;
  ctaDownloadTo?: string;
  ctaSignInLabel?: string;
  ctaSignInHint?: string;
  ctaSignInTo?: string;
  titleId?: string;
  variant?: "default" | "lead" | "compact";
  accentLine?: boolean;
};

export function PublicHeroSection({
  eyebrow,
  title,
  description,
  ctaDownloadLabel,
  ctaDownloadHint,
  ctaDownloadTo = "/download",
  ctaSignInLabel,
  ctaSignInHint,
  ctaSignInTo = "/login",
  titleId = "public-hero-title",
  variant = "default",
  accentLine = false,
}: PublicHeroSectionProps) {
  const heroClass =
    variant === "lead"
      ? "vyn-public-hero vyn-public-hero-lead"
      : variant === "compact"
        ? "vyn-public-hero vyn-public-hero-compact"
        : "vyn-public-hero";

  const showCtas = Boolean(ctaDownloadLabel || ctaSignInLabel);

  return (
    <div className={heroClass}>
      {accentLine ? <span className="vyn-public-hero-accent-line" aria-hidden="true" /> : null}
      <p className="vyn-public-eyebrow">{eyebrow}</p>
      <h1 id={titleId}>{title}</h1>
      <p className="vyn-public-hero-description">{description}</p>
      {showCtas ? (
        <div className="vyn-public-cta-stack">
          {ctaDownloadLabel ? (
            <div className="vyn-public-cta-item">
              <Link className="vyn-public-btn vyn-public-btn-primary" to={ctaDownloadTo}>
                {ctaDownloadLabel}
              </Link>
              {ctaDownloadHint ? <p className="vyn-public-cta-hint">{ctaDownloadHint}</p> : null}
            </div>
          ) : null}
          {ctaSignInLabel ? (
            <div className="vyn-public-cta-item">
              <Link className="vyn-public-btn vyn-public-btn-secondary" to={ctaSignInTo}>
                {ctaSignInLabel}
              </Link>
              {ctaSignInHint ? <p className="vyn-public-cta-hint">{ctaSignInHint}</p> : null}
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
