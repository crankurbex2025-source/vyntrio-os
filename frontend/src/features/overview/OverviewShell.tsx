import type { OverviewDto } from "./overviewDto";
import { formatOverviewCollectedAt } from "./overviewDto";

type OverviewShellProps = {
  overview: OverviewDto;
  isSigningOut: boolean;
  signOutError: boolean;
  settingsAccessError: boolean;
  settingsLoading: boolean;
  onOpenSettings: () => void;
  onSignOut: () => void;
};

export function OverviewShell({
  overview,
  isSigningOut,
  signOutError,
  settingsAccessError,
  settingsLoading,
  onOpenSettings,
  onSignOut,
}: OverviewShellProps) {
  const isReady = overview.readiness.status === "ready" && overview.readiness.database === "ok";

  return (
    <div className="dashboard-layout">
      <header className="dashboard-header">
        <div>
          <p className="dashboard-eyebrow">Vyntrio Control Center</p>
          <h1>{overview.instance.name}</h1>
          <p className="dashboard-subtitle">Read-only appliance overview</p>
        </div>
        <div className="dashboard-header-actions">
          <button
            type="button"
            className="dashboard-button dashboard-button-secondary"
            onClick={onOpenSettings}
            disabled={isSigningOut || settingsLoading}
          >
            {settingsLoading ? "Opening settings..." : "Instance settings"}
          </button>
          <button
            type="button"
            className="dashboard-button dashboard-button-primary"
            onClick={onSignOut}
            disabled={isSigningOut || settingsLoading}
          >
            {isSigningOut ? "Signing out..." : "Sign out"}
          </button>
        </div>
      </header>

      <main className="dashboard-main">
        <section className="dashboard-status-grid">
          <article
            className={`dashboard-status-card ${isReady ? "dashboard-status-card-ready" : "dashboard-status-card-not-ready"}`}
          >
            <p className="dashboard-card-label">Application readiness</p>
            <p className="dashboard-card-value">{isReady ? "Ready" : "Not ready"}</p>
            <p className="dashboard-card-detail">
              Database {overview.readiness.database === "ok" ? "connected" : "unavailable"}
            </p>
          </article>

          <article className="dashboard-info-card">
            <p className="dashboard-card-label">Service</p>
            <p className="dashboard-card-value">Running</p>
            <p className="dashboard-card-detail">API process is serving this overview</p>
          </article>
        </section>

        <section className="dashboard-panel">
          <h2>System information</h2>
          <dl className="dashboard-info-grid">
            <div className="dashboard-info-row">
              <dt>Instance</dt>
              <dd>{overview.instance.name}</dd>
            </div>
            <div className="dashboard-info-row">
              <dt>Version</dt>
              <dd>{overview.instance.version}</dd>
            </div>
            <div className="dashboard-info-row">
              <dt>Build</dt>
              <dd>{overview.instance.commit}</dd>
            </div>
            <div className="dashboard-info-row">
              <dt>Environment</dt>
              <dd>{overview.api.environment}</dd>
            </div>
            <div className="dashboard-info-row">
              <dt>Collected</dt>
              <dd>{formatOverviewCollectedAt(overview.collected_at)}</dd>
            </div>
          </dl>
        </section>

        {!isReady ? (
          <section className="dashboard-alert" role="status">
            The database is not ready. This overview reports current status only and does not
            perform recovery actions.
          </section>
        ) : null}

        {settingsAccessError ? (
          <section className="dashboard-alert" role="alert">
            You do not have access to instance settings.
          </section>
        ) : null}

        {signOutError ? (
          <section className="dashboard-alert" role="alert">
            Sign-out could not be completed. Please try again.
          </section>
        ) : null}
      </main>
    </div>
  );
}
