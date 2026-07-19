import type { ReactNode } from "react";

export type PublicHeroLayoutProps = {
  hero: ReactNode;
  companion?: ReactNode;
  art?: ReactNode;
  ariaLabelledBy?: string;
  premium?: boolean;
};

export function PublicHeroLayout({
  hero,
  companion,
  art,
  ariaLabelledBy = "public-hero-title",
  premium = false,
}: PublicHeroLayoutProps) {
  if (!premium) {
    return (
      <section className="vyn-public-hero-layout" aria-labelledby={ariaLabelledBy}>
        {hero}
        {companion}
      </section>
    );
  }

  return (
    <section
      className={[
        "vyn-public-hero-layout",
        "vyn-public-hero-layout-premium",
        companion ? "" : "vyn-public-hero-layout-solo",
      ]
        .filter(Boolean)
        .join(" ")}
      aria-labelledby={ariaLabelledBy}
    >
      <div className="vyn-public-hero-layout-copy">{hero}</div>
      {(art || companion) && (
        <div className="vyn-public-hero-layout-device">
          {art}
          {companion ? (
            <div className="vyn-public-hero-layout-device-frame">{companion}</div>
          ) : null}
        </div>
      )}
    </section>
  );
}
