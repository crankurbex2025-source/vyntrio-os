import type { ReactNode } from "react";

export type PublicHeroLayoutProps = {
  hero: ReactNode;
  companion: ReactNode;
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
      className="vyn-public-hero-layout vyn-public-hero-layout-premium"
      aria-labelledby={ariaLabelledBy}
    >
      <div className="vyn-public-hero-layout-copy">{hero}</div>
      <div className="vyn-public-hero-layout-device">
        {art}
        <div className="vyn-public-hero-layout-device-frame">{companion}</div>
      </div>
    </section>
  );
}
