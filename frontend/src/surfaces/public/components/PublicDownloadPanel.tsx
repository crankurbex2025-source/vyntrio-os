export type PublicDownloadRow = {
  label: string;
  value: string;
};

export type PublicDownloadPanelProps = {
  heading: string;
  intro?: string;
  rows: PublicDownloadRow[];
  note: string;
  headingId?: string;
};

export function PublicDownloadPanel({
  heading,
  intro,
  rows,
  note,
  headingId = "public-download-panel-heading",
}: PublicDownloadPanelProps) {
  return (
    <section className="vyn-public-download-panel" aria-labelledby={headingId}>
      <h2 id={headingId}>{heading}</h2>
      {intro ? <p className="vyn-public-download-panel-intro">{intro}</p> : null}
      <dl className="vyn-public-download-rows">
        {rows.map((row) => (
          <div key={row.label} className="vyn-public-download-row">
            <dt>{row.label}</dt>
            <dd>{row.value}</dd>
          </div>
        ))}
      </dl>
      <p className="vyn-public-download-panel-note">{note}</p>
    </section>
  );
}
