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
  chassis?: boolean;
  bezel?: { powerLabel: string; linkLabel: string };
};

export function PublicControlSurfaceFrame({
  heading,
  subheading,
  panelLabel,
  panelNote,
  rows,
  headingId = "public-surface-heading",
  variant = "compact",
  chassis = false,
  bezel,
}: PublicControlSurfaceFrameProps) {
  const surfaceClass = [
    variant === "showcase" ? "vyn-public-surface vyn-public-surface-showcase-frame" : "vyn-public-surface",
    chassis ? "vyn-public-surface-chassis" : "",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <aside className={surfaceClass} aria-labelledby={headingId}>
      {bezel ? (
        <div className="vyn-public-surface-bezel" aria-hidden="true">
          <span className="vyn-public-surface-bezel-lamp vyn-public-surface-bezel-lamp-standby" />
          <span className="vyn-public-surface-bezel-power">{bezel.powerLabel}</span>
          <span className="vyn-public-surface-bezel-lamp vyn-public-surface-bezel-lamp-off" />
          <span className="vyn-public-surface-bezel-link">{bezel.linkLabel}</span>
        </div>
      ) : null}
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
