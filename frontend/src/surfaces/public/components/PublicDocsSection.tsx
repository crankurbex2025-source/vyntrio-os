import { PublicResourceList, type PublicResourceItem } from "./PublicResourceList";
import { PublicSectionIntro } from "./PublicSectionIntro";

export type PublicDocsSectionProps = {
  eyebrow?: string;
  heading: string;
  intro?: string;
  items: PublicResourceItem[];
  headingId?: string;
};

export function PublicDocsSection({
  eyebrow,
  heading,
  intro,
  items,
  headingId,
}: PublicDocsSectionProps) {
  return (
    <section className="vyn-public-docs-section" aria-labelledby={headingId}>
      <PublicSectionIntro eyebrow={eyebrow} heading={heading} description={intro} headingId={headingId} />
      <PublicResourceList items={items} headingId={headingId} />
    </section>
  );
}
