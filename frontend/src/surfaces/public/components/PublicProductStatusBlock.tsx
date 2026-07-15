export type PublicProductStatusBlockProps = {
  eyebrow: string;
  heading: string;
  body: string;
  points: string[];
  headingId?: string;
  variant?: "default" | "terminal";
};

export function PublicProductStatusBlock({
  eyebrow,
  heading,
  body,
  points,
  headingId = "public-product-status-heading",
  variant = "default",
}: PublicProductStatusBlockProps) {
  const sectionClass =
    variant === "terminal"
      ? "vyn-public-product-status vyn-public-product-status-terminal"
      : "vyn-public-product-status";

  return (
    <section className={sectionClass} aria-labelledby={headingId}>
      <p className="vyn-public-eyebrow">{eyebrow}</p>
      <h2 id={headingId}>{heading}</h2>
      <p className="vyn-public-product-status-body">{body}</p>
      <ul className="vyn-public-product-status-list">
        {points.map((point) => (
          <li key={point}>{point}</li>
        ))}
      </ul>
    </section>
  );
}
