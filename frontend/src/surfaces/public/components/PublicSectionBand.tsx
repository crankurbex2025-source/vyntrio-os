import type { ReactNode } from "react";

export type PublicSectionBandProps = {
  children: ReactNode;
  tone?: "default" | "elevated" | "inset";
  surface?: "default" | "hero" | "statement" | "finale" | "capabilities" | "journey" | "route-hero" | "artifact";
  id?: string;
};

export function PublicSectionBand({
  children,
  tone = "default",
  surface = "default",
  id,
}: PublicSectionBandProps) {
  const toneClass =
    tone === "elevated"
      ? "vyn-public-section-band vyn-public-section-band-elevated"
      : tone === "inset"
        ? "vyn-public-section-band vyn-public-section-band-inset"
        : "vyn-public-section-band";

  const surfaceClass =
    surface === "default" ? "" : ` vyn-public-section-band-surface-${surface}`;

  return (
    <div id={id} className={`${toneClass}${surfaceClass}`}>
      <div className="vyn-public-section-band-inner vyn-public-container-root">{children}</div>
    </div>
  );
}
