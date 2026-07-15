import { Link } from "react-router-dom";

export type PublicFinalCtaBandProps = {
  heading: string;
  body: string;
  ctaDownloadLabel: string;
  ctaDownloadTo?: string;
  ctaSignInLabel: string;
  ctaSignInTo?: string;
  headingId?: string;
};

export function PublicFinalCtaBand({
  heading,
  body,
  ctaDownloadLabel,
  ctaDownloadTo = "/download",
  ctaSignInLabel,
  ctaSignInTo = "/login",
  headingId = "public-final-cta-heading",
}: PublicFinalCtaBandProps) {
  return (
    <section className="vyn-public-final-cta" aria-labelledby={headingId}>
      <h2 id={headingId}>{heading}</h2>
      <p>{body}</p>
      <div className="vyn-public-cta-row">
        <Link className="vyn-public-btn vyn-public-btn-primary" to={ctaDownloadTo}>
          {ctaDownloadLabel}
        </Link>
        <Link className="vyn-public-btn vyn-public-btn-secondary" to={ctaSignInTo}>
          {ctaSignInLabel}
        </Link>
      </div>
    </section>
  );
}
