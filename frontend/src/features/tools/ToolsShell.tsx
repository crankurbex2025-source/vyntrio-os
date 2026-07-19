import { AppliancePageHeader } from "../../surfaces/appliance/ui/AppliancePageHeader";
import { AppliancePanel } from "../../surfaces/appliance/ui/AppliancePanel";

export function ToolsShell() {
  return (
    <div className="vyn-ops-page">
      <AppliancePageHeader
        title="Tools"
        status="System utilities scaffold · events and logs APIs are not wired yet"
      />

      <AppliancePanel
        title="Events"
        planned
        note="System event feed and notification history will appear here once a backend events API exists."
      >
        <p className="vyn-ops-empty">Not wired — Planned. No event feed is available today.</p>
      </AppliancePanel>

      <AppliancePanel
        title="Logs"
        planned
        note="Application and system log viewers require a logs API."
      >
        <p className="vyn-ops-empty">Not wired — Planned. No log viewer is available today.</p>
      </AppliancePanel>

      <AppliancePanel
        title="Diagnostics"
        planned
        note="Support bundles and scheduled admin tasks (parity/mover analogs) stay reserved until implemented."
      >
        <p className="vyn-ops-empty">Not wired — Planned. No diagnostics tools are available today.</p>
      </AppliancePanel>
    </div>
  );
}
