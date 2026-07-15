import type { ReactNode } from "react";

export type PublicSectionBandProps = {
  children: ReactNode;
  tone?: "default" | "elevated" | "inset";
  id?: string;
};

export function PublicSectionBand({ children, tone = "default", id }: PublicSectionBandProps) {
  return (
    <div
      id={id}
      className={
        tone === "elevated"
          ? "vyn-public-section-band vyn-public-section-band-elevated"
          : tone === "inset"
            ? "vyn-public-section-band vyn-public-section-band-inset"
            : "vyn-public-section-band"
      }
    >
      <div className="vyn-public-section-band-inner">{children}</div>
    </div>
  );
}
