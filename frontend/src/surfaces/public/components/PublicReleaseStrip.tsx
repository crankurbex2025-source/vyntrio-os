export type PublicReleaseStripProps = {
  label: string;
  title: string;
  body: string;
  titleId?: string;
};

export function PublicReleaseStrip({
  label,
  title,
  body,
  titleId = "public-release-title",
}: PublicReleaseStripProps) {
  return (
    <section className="vyn-public-release" aria-labelledby={titleId}>
      <p className="vyn-public-release-label">{label}</p>
      <h2 id={titleId}>{title}</h2>
      <p>{body}</p>
    </section>
  );
}
