export type PublicPillarGlyphKind = "storage" | "services" | "control";

export type PublicPillarGlyphProps = {
  kind: PublicPillarGlyphKind;
};

export function PublicPillarGlyph({ kind }: PublicPillarGlyphProps) {
  if (kind === "storage") {
    return (
      <svg
        className="vyn-public-pillar-glyph"
        viewBox="0 0 40 40"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        aria-hidden="true"
        focusable="false"
      >
        <rect x="6" y="8" width="28" height="24" rx="4" stroke="currentColor" strokeOpacity="0.35" />
        <rect x="11" y="13" width="18" height="5" rx="1.5" stroke="currentColor" strokeOpacity="0.22" />
        <rect x="11" y="21" width="18" height="5" rx="1.5" stroke="currentColor" strokeOpacity="0.18" />
        <circle cx="14" cy="15.5" r="1.2" fill="#e85d2b" fillOpacity="0.8" />
      </svg>
    );
  }

  if (kind === "services") {
    return (
      <svg
        className="vyn-public-pillar-glyph"
        viewBox="0 0 40 40"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        aria-hidden="true"
        focusable="false"
      >
        <rect x="8" y="10" width="10" height="8" rx="2" stroke="currentColor" strokeOpacity="0.28" />
        <rect x="22" y="10" width="10" height="8" rx="2" stroke="currentColor" strokeOpacity="0.22" />
        <rect x="8" y="22" width="10" height="8" rx="2" stroke="currentColor" strokeOpacity="0.22" />
        <rect x="22" y="22" width="10" height="8" rx="2" stroke="currentColor" strokeOpacity="0.18" />
        <path d="M18 14h4M18 26h4" stroke="currentColor" strokeOpacity="0.2" strokeLinecap="round" />
      </svg>
    );
  }

  return (
    <svg
      className="vyn-public-pillar-glyph"
      viewBox="0 0 40 40"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
      focusable="false"
    >
      <rect x="7" y="9" width="26" height="22" rx="4" stroke="currentColor" strokeOpacity="0.32" />
      <rect x="11" y="13" width="18" height="10" rx="2" stroke="currentColor" strokeOpacity="0.2" />
      <line x1="11" y1="27" x2="29" y2="27" stroke="currentColor" strokeOpacity="0.16" />
      <circle cx="14" cy="27" r="1.2" fill="currentColor" fillOpacity="0.35" />
      <circle cx="18" cy="27" r="1.2" fill="#e85d2b" fillOpacity="0.75" />
    </svg>
  );
}
