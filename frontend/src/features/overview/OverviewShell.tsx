import { Link } from "react-router-dom";
import { AppliancePageHeader } from "../../surfaces/appliance/ui/AppliancePageHeader";
import { AppliancePanel } from "../../surfaces/appliance/ui/AppliancePanel";
import type { OverviewDto } from "./overviewDto";
import {
  formatBackupFailureDetail,
  formatMetricBytes,
  formatOverviewCollectedAt,
} from "./overviewDto";

type OverviewShellProps = {
  overview: OverviewDto;
  signOutError: boolean;
  settingsAccessError: boolean;
  storageAccessError: boolean;
  settingsLoading: boolean;
  storageLoading: boolean;
};

export function OverviewShell({
  overview,
  signOutError,
  settingsAccessError,
  storageAccessError,
  settingsLoading,
  storageLoading,
}: OverviewShellProps) {
  const stateFilesystem = overview.host.filesystems[0];
  const runtimeLabel =
    overview.runtime.status === "ready"
      ? "Ready"
      : overview.runtime.status === "degraded"
        ? "Degraded"
        : "Unknown";

  const cpuSummary =
    overview.host.cpu.status === "ok"
      ? `${overview.host.cpu.logical_cores} cores · load ${overview.host.cpu.load_1m?.toFixed(2)}`
      : "Unavailable";
  const memSummary =
    overview.host.memory.status === "ok"
      ? `${formatMetricBytes(overview.host.memory.used_bytes ?? 0)} / ${formatMetricBytes(overview.host.memory.total_bytes ?? 0)}`
      : "Unavailable";
  const storageSummary =
    overview.storage.status === "ok"
      ? `${overview.storage.disk_count} disks · ${overview.storage.pool_count} pools · ${overview.storage.share_count} plans`
      : "Unavailable";
  const backupSummary =
    overview.backup.status === "never_run"
      ? "Never run"
      : overview.backup.status === "succeeded"
        ? "Last succeeded"
        : overview.backup.status === "failed"
          ? "Last failed"
          : "Unavailable";
  const networkSummary =
    overview.network.status === "available"
      ? "Interface present"
      : overview.network.status === "unknown"
        ? "Unclear"
        : "Unavailable";
  const healthSummary =
    overview.health.status === "healthy"
      ? "Healthy"
      : overview.health.status === "warning"
        ? "Warning"
        : "Unknown";
  const softwareSummary =
    overview.software.status === "ok"
      ? overview.software.version
      : "Unavailable";

  return (
    <div className="vyn-ops-page">
      <AppliancePageHeader
        title="Dashboard"
        status={`Read-only overview · collected ${formatOverviewCollectedAt(overview.collected_at)}`}
      />

      <section className="vyn-ops-metric-strip" aria-label="System snapshot">
        <div className="vyn-ops-metric">
          <p className="vyn-ops-metric-label">Runtime</p>
          <p className="vyn-ops-metric-value">{runtimeLabel}</p>
          <p className="vyn-ops-metric-detail">
            {overview.runtime.status === "degraded" && overview.runtime.note === "database"
              ? "Database not ready"
              : "Point-in-time"}
          </p>
        </div>
        <div className="vyn-ops-metric">
          <p className="vyn-ops-metric-label">Instance</p>
          <p className="vyn-ops-metric-value">{overview.instance.name}</p>
        </div>
        <div className="vyn-ops-metric">
          <p className="vyn-ops-metric-label">CPU</p>
          <p className="vyn-ops-metric-value">{cpuSummary}</p>
        </div>
        <div className="vyn-ops-metric">
          <p className="vyn-ops-metric-label">Memory</p>
          <p className="vyn-ops-metric-value">{memSummary}</p>
        </div>
        <div className="vyn-ops-metric">
          <p className="vyn-ops-metric-label">Storage</p>
          <p className="vyn-ops-metric-value">{storageSummary}</p>
          <p className="vyn-ops-metric-detail">Format not applied · shares not published</p>
        </div>
        <div className="vyn-ops-metric">
          <p className="vyn-ops-metric-label">Backup</p>
          <p className="vyn-ops-metric-value">{backupSummary}</p>
        </div>
        <div className="vyn-ops-metric">
          <p className="vyn-ops-metric-label">Network</p>
          <p className="vyn-ops-metric-value">{networkSummary}</p>
          <p className="vyn-ops-metric-detail">No WAN/DNS claim</p>
        </div>
        <div className="vyn-ops-metric">
          <p className="vyn-ops-metric-label">Health</p>
          <p className="vyn-ops-metric-value">{healthSummary}</p>
        </div>
      </section>

      <section className="vyn-ops-setup-banner" aria-label="Setup progress">
        <strong>Setup</strong>
        <ol>
          <li>
            <Link to="/app/storage">
              {storageLoading ? "Storage (loading…)" : "Review disks / declare pool"}
            </Link>
          </li>
          <li>
            <Link to="/app/shares">Prepare share plan</Link>
          </li>
          <li>
            <Link to="/app/settings">
              {settingsLoading ? "Settings (loading…)" : "Instance name"}
            </Link>
          </li>
        </ol>
      </section>

      {overview.readiness.status !== "ready" ? (
        <section className="vyn-ops-alert" role="status">
          Database is not ready. This overview reports status only and does not perform recovery.
        </section>
      ) : null}
      {settingsAccessError ? (
        <section className="vyn-ops-alert" role="alert">
          You do not have access to instance settings.
        </section>
      ) : null}
      {storageAccessError ? (
        <section className="vyn-ops-alert" role="alert">
          You do not have access to the storage inventory.
        </section>
      ) : null}
      {signOutError ? (
        <section className="vyn-ops-alert" role="alert">
          Sign-out could not be completed. Please try again.
        </section>
      ) : null}

      <div className="vyn-ops-grid-2">
        <AppliancePanel
          title="Host"
          note="Point-in-time snapshot. Load average is not CPU utilization."
        >
          <dl className="vyn-ops-dl">
            <div className="vyn-ops-dl-row">
              <dt>CPU</dt>
              <dd>
                {overview.host.cpu.status === "ok"
                  ? `${overview.host.cpu.logical_cores} cores · 1-minute load ${overview.host.cpu.load_1m?.toFixed(2)}`
                  : "CPU metrics could not be collected"}
              </dd>
            </div>
            <div className="vyn-ops-dl-row">
              <dt>Memory</dt>
              <dd>
                {overview.host.memory.status === "ok"
                  ? `${formatMetricBytes(overview.host.memory.used_bytes ?? 0)} used · ${formatMetricBytes(overview.host.memory.available_bytes ?? 0)} available of ${formatMetricBytes(overview.host.memory.total_bytes ?? 0)}`
                  : "Memory metrics could not be collected"}
              </dd>
            </div>
            <div className="vyn-ops-dl-row">
              <dt>State FS</dt>
              <dd>
                {stateFilesystem.status === "ok"
                  ? `${formatMetricBytes(stateFilesystem.used_bytes ?? 0)} used · ${formatMetricBytes(stateFilesystem.available_bytes ?? 0)} available${stateFilesystem.fs_type ? ` · ${stateFilesystem.fs_type}` : ""}`
                  : "Storage metrics could not be collected"}
              </dd>
            </div>
          </dl>
        </AppliancePanel>

        <AppliancePanel
          title="Storage layout"
          note="Inventory counts only. Declared pools and share plans are state — format and SMB/NFS are not live."
        >
          <dl className="vyn-ops-dl">
            <div className="vyn-ops-dl-row">
              <dt>Block devices</dt>
              <dd>
                {overview.storage.status === "ok"
                  ? `${overview.storage.disk_count} reported · ${overview.storage.eligible_count} eligible · ${overview.storage.excluded_count} excluded`
                  : "Storage inventory could not be summarized for this overview."}
              </dd>
            </div>
            <div className="vyn-ops-dl-row">
              <dt>Pools</dt>
              <dd>{overview.storage.pool_count} declared · on-disk format not applied</dd>
            </div>
            <div className="vyn-ops-dl-row">
              <dt>Shares</dt>
              <dd>{overview.storage.share_count} planned · protocols not published</dd>
            </div>
          </dl>
        </AppliancePanel>

        <AppliancePanel
          title="Software"
          note="Metadata from the running API. Does not check for updates."
        >
          <dl className="vyn-ops-dl">
            <div className="vyn-ops-dl-row">
              <dt>Version</dt>
              <dd>{softwareSummary}</dd>
            </div>
            <div className="vyn-ops-dl-row">
              <dt>Build</dt>
              <dd>
                {overview.software.status === "ok"
                  ? [
                      overview.software.commit
                        ? `Build ${overview.software.commit}`
                        : "Build revision not recorded",
                      overview.software.channel === "development"
                        ? "development channel"
                        : overview.software.channel === "production"
                          ? "production channel"
                          : overview.software.channel === "unknown"
                            ? "channel unknown"
                            : null,
                    ]
                      .filter(Boolean)
                      .join(" · ")
                  : "Software release metadata could not be determined."}
              </dd>
            </div>
            <div className="vyn-ops-dl-row">
              <dt>Environment</dt>
              <dd>{overview.api.environment}</dd>
            </div>
          </dl>
        </AppliancePanel>

        <AppliancePanel
          title="Health and backup"
          note="Derived from overview slices only — not a full appliance health probe."
        >
          <dl className="vyn-ops-dl">
            <div className="vyn-ops-dl-row">
              <dt>Health</dt>
              <dd>
                {overview.health.status === "healthy"
                  ? `Healthy · ${formatOverviewCollectedAt(overview.collected_at)}`
                  : overview.health.status === "warning"
                    ? overview.health.note === "database"
                      ? "Warning · database dependency is not ready."
                      : overview.health.note === "backup"
                        ? "Warning · last recorded local backup attempt failed."
                        : "Warning · a recorded overview signal needs attention."
                    : "Health summary could not be classified from current overview state."}
              </dd>
            </div>
            <div className="vyn-ops-dl-row">
              <dt>Backup</dt>
              <dd>
                {overview.backup.status === "never_run"
                  ? "No completed local backup has been recorded yet."
                  : overview.backup.status === "succeeded"
                    ? `Last succeeded · ${formatOverviewCollectedAt(overview.backup.completed_at ?? "")}`
                    : overview.backup.status === "failed"
                      ? `Last failed · ${formatOverviewCollectedAt(overview.backup.completed_at ?? "")}${overview.backup.failure ? ` · ${formatBackupFailureDetail(overview.backup.failure)}` : ""}`
                      : "Backup status could not be read."}
              </dd>
            </div>
            <div className="vyn-ops-dl-row">
              <dt>Network</dt>
              <dd>
                {overview.network.status === "available"
                  ? "Local non-loopback interface present. Does not verify internet, DNS, or reachability."
                  : overview.network.status === "unknown"
                    ? "No eligible interface observed from this process. Does not prove hardware is missing."
                    : "Network presence could not be determined."}
              </dd>
            </div>
          </dl>
        </AppliancePanel>
      </div>
    </div>
  );
}
