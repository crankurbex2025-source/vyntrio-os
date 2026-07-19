import type { ReactNode } from "react";

export type PublicSectionBandProps = {
  children: ReactNode;
  tone?: "default" | "elevated" | "inset" | "trust" | "panel";
  surface?:
    | "default"
    | "hero"
    | "statement"
    | "finale"
    | "capabilities"
    | "journey"
    | "route-hero"
    | "artifact"
    | "trust"
    | "enterprise";
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
        : tone === "trust"
          ? "vyn-public-section-band vyn-public-section-band-trust"
          : tone === "panel"
            ? "vyn-public-section-band vyn-public-section-band-panel"
            : "vyn-public-section-band";

  const surfaceClass =
    surface === "default" ? "" : ` vyn-public-section-band-surface-${surface}`;

  return (
    <div id={id} className={`${toneClass}${surfaceClass}`}>
      <div className="vyn-public-section-band-inner vyn-public-container-root">{children}</div>
    </div>
  );
}

