import type { ReactNode } from "react";

type AppliancePageHeaderProps = {
  title: string;
  status?: string;
  actions?: ReactNode;
};

export function AppliancePageHeader({ title, status, actions }: AppliancePageHeaderProps) {
  return (
    <header className="vyn-ops-page-header">
      <div>
        <h1>{title}</h1>
        {status ? <p className="vyn-ops-page-status">{status}</p> : null}
      </div>
      {actions ? <div className="vyn-ops-page-actions">{actions}</div> : null}
    </header>
  );
}
