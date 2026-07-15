import { Link } from "react-router-dom";

export type PublicInlineCtaBandProps = {
  heading: string;
  body?: string;
  primaryLabel: string;
  primaryTo?: string;
  secondaryLabel?: string;
  secondaryTo?: string;
  headingId?: string;
};

export function PublicInlineCtaBand({
  heading,
  body,
  primaryLabel,
  primaryTo = "/design-preview/landing",
  secondaryLabel,
  secondaryTo,
  headingId = "public-inline-cta-heading",
}: PublicInlineCtaBandProps) {
  return (
    <section className="vyn-public-inline-cta" aria-labelledby={headingId}>
      <h2 id={headingId}>{heading}</h2>
      {body ? <p>{body}</p> : null}
      <div className="vyn-public-cta-row">
        <Link className="vyn-public-btn vyn-public-btn-primary" to={primaryTo}>
          {primaryLabel}
        </Link>
        {secondaryLabel && secondaryTo ? (
          <Link className="vyn-public-btn vyn-public-btn-secondary" to={secondaryTo}>
            {secondaryLabel}
          </Link>
        ) : null}
      </div>
    </section>
  );
}
