export type PublicUseCase = {
  tag?: string;
  title: string;
  body: string;
};

export type PublicUseCaseSectionProps = {
  eyebrow?: string;
  heading: string;
  intro?: string;
  ariaLabel?: string;
  cases: PublicUseCase[];
  headingId?: string;
};

export function PublicUseCaseSection({
  eyebrow,
  heading,
  intro,
  ariaLabel,
  cases,
  headingId = "public-use-cases-heading",
}: PublicUseCaseSectionProps) {
  return (
    <section
      className="vyn-public-pillar-section vyn-public-use-cases"
      aria-labelledby={headingId}
    >
      {eyebrow ? <p className="vyn-public-eyebrow">{eyebrow}</p> : null}
      <h2 id={headingId} className="vyn-public-section-title">
        {heading}
      </h2>
      {intro ? <p className="vyn-public-pillar-intro">{intro}</p> : null}
      <div className="vyn-public-pillars" role="list" aria-label={ariaLabel}>
        {cases.map((useCase) => (
          <article key={useCase.title} className="vyn-public-pillar" role="listitem">
            {useCase.tag ? <p className="vyn-public-pillar-tag">{useCase.tag}</p> : null}
            <h3>{useCase.title}</h3>
            <p>{useCase.body}</p>
          </article>
        ))}
      </div>
    </section>
  );
}
