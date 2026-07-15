import type { ReactNode } from "react";

export type PublicHeroLayoutProps = {
  hero: ReactNode;
  companion: ReactNode;
  ariaLabelledBy?: string;
};

export function PublicHeroLayout({
  hero,
  companion,
  ariaLabelledBy = "public-hero-title",
}: PublicHeroLayoutProps) {
  return (
    <section className="vyn-public-hero-layout" aria-labelledby={ariaLabelledBy}>
      {hero}
      {companion}
    </section>
  );
}
