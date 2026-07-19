import "./PublicTrustBand.css";

export type PublicTrustMark = {
  label: string;
  detail?: string;
};

export type PublicTrustBandProps = {
  heading: string;
  headingEmphasis?: string;
  intro?: string;
  marks: PublicTrustMark[];
  headingId?: string;
};

/** Original Vyntrio trust/ecosystem band — text marks only, no third-party logos. */
export function PublicTrustBand({
  heading,
  headingEmphasis,
  intro,
  marks,
  headingId = "public-trust-band-heading",
}: PublicTrustBandProps) {
  return (
    <section className="vyn-public-trust-band" aria-labelledby={headingId}>
      <h2 id={headingId} className="vyn-public-trust-band-heading">
        {heading}
        {headingEmphasis ? (
          <>
            {" "}
            <span className="vyn-public-trust-band-emphasis">{headingEmphasis}</span>
          </>
        ) : null}
      </h2>
      {intro ? <p className="vyn-public-trust-band-intro">{intro}</p> : null}
      <ul className="vyn-public-trust-band-marks">
        {marks.map((mark) => (
          <li key={mark.label} className="vyn-public-trust-band-mark">
            <span className="vyn-public-trust-band-mark-label">{mark.label}</span>
            {mark.detail ? (
              <span className="vyn-public-trust-band-mark-detail">{mark.detail}</span>
            ) : null}
          </li>
        ))}
      </ul>
    </section>
  );
}
