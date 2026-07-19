import type { ReactNode } from "react";

type AppliancePanelProps = {
  title: string;
  note?: string;
  children: ReactNode;
  planned?: boolean;
};

export function AppliancePanel({ title, note, children, planned = false }: AppliancePanelProps) {
  return (
    <section className={`vyn-ops-panel${planned ? " vyn-ops-panel-planned" : ""}`}>
      <div className="vyn-ops-panel-head">
        <h2>{title}</h2>
        {planned ? <span className="vyn-ops-planned-badge">Planned</span> : null}
      </div>
      {note ? <p className="vyn-ops-panel-note">{note}</p> : null}
      <div className="vyn-ops-panel-body">{children}</div>
    </section>
  );
}
