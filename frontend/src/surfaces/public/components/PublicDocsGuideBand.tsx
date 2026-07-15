import { PublicSectionIntro } from "./PublicSectionIntro";

export type PublicDocsGuideBandProps = {
  eyebrow?: string;
  heading: string;
  body: string;
  headingId?: string;
};

export function PublicDocsGuideBand({
  eyebrow,
  heading,
  body,
  headingId = "public-docs-guide-heading",
}: PublicDocsGuideBandProps) {
  return (
    <section className="vyn-public-docs-guide" aria-labelledby={headingId}>
      <PublicSectionIntro eyebrow={eyebrow} heading={heading} description={body} headingId={headingId} />
    </section>
  );
}
