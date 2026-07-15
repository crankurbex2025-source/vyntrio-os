import { Link } from "react-router-dom";

export type PublicHeroSectionProps = {
  eyebrow: string;
  title: string;
  description: string;
  ctaDownloadLabel: string;
  ctaDownloadHint?: string;
  ctaDownloadTo?: string;
  ctaSignInLabel: string;
  ctaSignInHint?: string;
  ctaSignInTo?: string;
  titleId?: string;
  variant?: "default" | "lead";
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
}: PublicHeroSectionProps) {
  const heroClass =
    variant === "lead" ? "vyn-public-hero vyn-public-hero-lead" : "vyn-public-hero";

  return (
    <div className={heroClass}>
      <p className="vyn-public-eyebrow">{eyebrow}</p>
      <h1 id={titleId}>{title}</h1>
      <p className="vyn-public-hero-description">{description}</p>
      <div className="vyn-public-cta-stack">
        <div className="vyn-public-cta-item">
          <Link className="vyn-public-btn vyn-public-btn-primary" to={ctaDownloadTo}>
            {ctaDownloadLabel}
          </Link>
          {ctaDownloadHint ? <p className="vyn-public-cta-hint">{ctaDownloadHint}</p> : null}
        </div>
        <div className="vyn-public-cta-item">
          <Link className="vyn-public-btn vyn-public-btn-secondary" to={ctaSignInTo}>
            {ctaSignInLabel}
          </Link>
          {ctaSignInHint ? <p className="vyn-public-cta-hint">{ctaSignInHint}</p> : null}
        </div>
      </div>
    </div>
  );
}
