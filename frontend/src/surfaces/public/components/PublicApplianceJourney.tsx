export type PublicApplianceJourneyStep = {
  phase: string;
  title: string;
  body: string;
  status: string;
};

export type PublicApplianceJourneyProps = {
  eyebrow?: string;
  heading: string;
  intro?: string;
  steps: PublicApplianceJourneyStep[];
  headingId?: string;
  ariaLabel: string;
};

export function PublicApplianceJourney({
  eyebrow,
  heading,
  intro,
  steps,
  headingId = "public-appliance-journey-heading",
  ariaLabel,
}: PublicApplianceJourneyProps) {
  return (
    <section className="vyn-public-appliance-journey" aria-labelledby={headingId}>
      {eyebrow ? <p className="vyn-public-eyebrow">{eyebrow}</p> : null}
      <h2 id={headingId} className="vyn-public-section-title">
        {heading}
      </h2>
      {intro ? <p className="vyn-public-journey-intro">{intro}</p> : null}
      <ol className="vyn-public-journey-steps" aria-label={ariaLabel}>
        {steps.map((step) => (
          <li key={step.phase} className="vyn-public-journey-step">
            <span className="vyn-public-journey-phase">{step.phase}</span>
            <div className="vyn-public-journey-copy">
              <div className="vyn-public-journey-step-header">
                <h3>{step.title}</h3>
                <span className="vyn-public-journey-status">{step.status}</span>
              </div>
              <p>{step.body}</p>
            </div>
          </li>
        ))}
      </ol>
    </section>
  );
}
