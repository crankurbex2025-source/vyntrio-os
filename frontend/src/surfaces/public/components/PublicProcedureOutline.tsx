import { PublicSectionIntro } from "./PublicSectionIntro";

export type PublicProcedureStep = {
  step: string;
  title: string;
  body: string;
};

export type PublicProcedureOutlineProps = {
  eyebrow?: string;
  heading: string;
  intro?: string;
  steps: PublicProcedureStep[];
  headingId?: string;
  ariaLabel: string;
};

export function PublicProcedureOutline({
  eyebrow,
  heading,
  intro,
  steps,
  headingId = "public-procedure-outline-heading",
  ariaLabel,
}: PublicProcedureOutlineProps) {
  return (
    <section className="vyn-public-procedure-outline" aria-labelledby={headingId}>
      <PublicSectionIntro eyebrow={eyebrow} heading={heading} description={intro} headingId={headingId} />
      <ol className="vyn-public-procedure-steps" aria-label={ariaLabel}>
        {steps.map((step) => (
          <li key={step.step} className="vyn-public-procedure-step">
            <span className="vyn-public-procedure-index">{step.step}</span>
            <div className="vyn-public-procedure-copy">
              <h3>{step.title}</h3>
              <p>{step.body}</p>
            </div>
          </li>
        ))}
      </ol>
    </section>
  );
}
