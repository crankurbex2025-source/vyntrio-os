export type PublicApplianceSignalStep = {
  label: string;
  detail: string;
};

export type PublicApplianceSignalPathProps = {
  steps: PublicApplianceSignalStep[];
  ariaLabel: string;
};

export function PublicApplianceSignalPath({ steps, ariaLabel }: PublicApplianceSignalPathProps) {
  return (
    <ol className="vyn-public-signal-path" aria-label={ariaLabel}>
      {steps.map((step, index) => (
        <li key={step.label} className="vyn-public-signal-path-step">
          <span className="vyn-public-signal-path-index">
            {String(index + 1).padStart(2, "0")}
          </span>
          <span className="vyn-public-signal-path-copy">
            <span className="vyn-public-signal-path-label">{step.label}</span>
            <span className="vyn-public-signal-path-detail">{step.detail}</span>
          </span>
          {index < steps.length - 1 ? (
            <span className="vyn-public-signal-path-connector" aria-hidden="true" />
          ) : null}
        </li>
      ))}
    </ol>
  );
}
