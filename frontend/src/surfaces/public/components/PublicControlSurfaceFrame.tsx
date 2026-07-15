export type PublicSurfaceRow = {
  label: string;
  value: string;
};

export type PublicControlSurfaceFrameProps = {
  heading: string;
  subheading: string;
  panelLabel: string;
  panelNote: string;
  rows: PublicSurfaceRow[];
  headingId?: string;
  variant?: "compact" | "showcase";
};

export function PublicControlSurfaceFrame({
  heading,
  subheading,
  panelLabel,
  panelNote,
  rows,
  headingId = "public-surface-heading",
  variant = "compact",
}: PublicControlSurfaceFrameProps) {
  const surfaceClass =
    variant === "showcase"
      ? "vyn-public-surface vyn-public-surface-showcase-frame"
      : "vyn-public-surface";

  return (
    <aside className={surfaceClass} aria-labelledby={headingId}>
      <div className="vyn-public-surface-header">
        <h2 id={headingId}>{heading}</h2>
        <p>{subheading}</p>
      </div>
      <div className="vyn-public-surface-panel">
        <p className="vyn-public-surface-panel-label">{panelLabel}</p>
        <dl className="vyn-public-surface-rows">
          {rows.map((row) => (
            <div key={row.label} className="vyn-public-surface-row">
              <dt>{row.label}</dt>
              <dd>{row.value}</dd>
            </div>
          ))}
        </dl>
        <p className="vyn-public-surface-note">{panelNote}</p>
      </div>
    </aside>
  );
}
