import { PublicPillarGlyph, type PublicPillarGlyphKind } from "./PublicPillarGlyph";

export type PublicPillar = {
  tag?: string;
  title: string;
  body: string;
  featured?: boolean;
  glyph?: PublicPillarGlyphKind;
};

export type PublicPillarSectionProps = {
  eyebrow?: string;
  heading: string;
  intro?: string;
  pillars: PublicPillar[];
  headingId?: string;
  showGlyphs?: boolean;
};

export function PublicPillarSection({
  eyebrow,
  heading,
  intro,
  pillars,
  headingId = "public-pillars-heading",
  showGlyphs = false,
}: PublicPillarSectionProps) {
  return (
    <section className="vyn-public-pillar-section" aria-labelledby={headingId}>
      {eyebrow ? <p className="vyn-public-eyebrow">{eyebrow}</p> : null}
      <h2 id={headingId} className="vyn-public-section-title">
        {heading}
      </h2>
      {intro ? <p className="vyn-public-pillar-intro">{intro}</p> : null}
      <div className="vyn-public-pillars">
        {pillars.map((pillar) => (
          <article
            key={pillar.title}
            className={
              pillar.featured ? "vyn-public-pillar vyn-public-pillar-featured" : "vyn-public-pillar"
            }
          >
            {showGlyphs && pillar.glyph ? (
              <div className="vyn-public-pillar-glyph-wrap">
                <PublicPillarGlyph kind={pillar.glyph} />
              </div>
            ) : null}
            {pillar.tag ? <p className="vyn-public-pillar-tag">{pillar.tag}</p> : null}
            <h3>{pillar.title}</h3>
            <p>{pillar.body}</p>
          </article>
        ))}
      </div>
    </section>
  );
}
