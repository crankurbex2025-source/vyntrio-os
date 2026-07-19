export type PublicAppsVmsSectionProps = {
  eyebrow: string;
  heading: string;
  intro: string;
  appsTitle: string;
  appsBody: string;
  appsStatus: string;
  vmsTitle: string;
  vmsBody: string;
  vmsStatus: string;
  headingId?: string;
};

export function PublicAppsVmsSection({
  eyebrow,
  heading,
  intro,
  appsTitle,
  appsBody,
  appsStatus,
  vmsTitle,
  vmsBody,
  vmsStatus,
  headingId = "public-apps-vms-heading",
}: PublicAppsVmsSectionProps) {
  return (
    <section className="vyn-public-apps-vms" aria-labelledby={headingId}>
      <p className="vyn-public-section-eyebrow">{eyebrow}</p>
      <h2 id={headingId}>{heading}</h2>
      <p className="vyn-public-download-panel-intro">{intro}</p>
      <div className="vyn-public-pillars">
        <article className="vyn-public-pillar">
          <p className="vyn-public-pillar-tag">{appsStatus}</p>
          <h3>{appsTitle}</h3>
          <p>{appsBody}</p>
        </article>
        <article className="vyn-public-pillar">
          <p className="vyn-public-pillar-tag">{vmsStatus}</p>
          <h3>{vmsTitle}</h3>
          <p>{vmsBody}</p>
        </article>
      </div>
    </section>
  );
}
