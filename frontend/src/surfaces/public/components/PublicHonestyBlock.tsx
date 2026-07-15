export type PublicHonestyBlockProps = {
  heading: string;
  body: string;
  headingId?: string;
};

export function PublicHonestyBlock({
  heading,
  body,
  headingId = "public-honesty-heading",
}: PublicHonestyBlockProps) {
  return (
    <section className="vyn-public-honesty" aria-labelledby={headingId}>
      <h2 id={headingId}>{heading}</h2>
      <p>{body}</p>
    </section>
  );
}
