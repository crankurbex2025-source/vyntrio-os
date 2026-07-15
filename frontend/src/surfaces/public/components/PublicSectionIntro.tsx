export type PublicSectionIntroProps = {
  eyebrow?: string;
  heading: string;
  description?: string;
  headingId?: string;
};

export function PublicSectionIntro({
  eyebrow,
  heading,
  description,
  headingId,
}: PublicSectionIntroProps) {
  return (
    <header className="vyn-public-section-intro">
      {eyebrow ? <p className="vyn-public-eyebrow">{eyebrow}</p> : null}
      <h2 id={headingId} className="vyn-public-section-intro-heading">
        {heading}
      </h2>
      {description ? <p className="vyn-public-section-intro-description">{description}</p> : null}
    </header>
  );
}
