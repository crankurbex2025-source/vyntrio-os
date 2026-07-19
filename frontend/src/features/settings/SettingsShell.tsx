import { AppliancePageHeader } from "../../surfaces/appliance/ui/AppliancePageHeader";
import { AppliancePanel } from "../../surfaces/appliance/ui/AppliancePanel";
import type { PublicSettingsDto } from "./settingsDto";

type SettingsShellProps = {
  settings: PublicSettingsDto;
  editMode: boolean;
  draftDisplayName: string;
  isUpdating: boolean;
  updateErrorMessage: string | null;
  updateValidationMessage: string | null;
  onStartEdit: () => void;
  onCancelEdit: () => void;
  onSaveDisplayName: () => void;
  onDraftDisplayNameChange: (value: string) => void;
  controlsLocked?: boolean;
};

export function SettingsShell({
  settings,
  editMode,
  draftDisplayName,
  isUpdating,
  updateErrorMessage,
  updateValidationMessage,
  onStartEdit,
  onCancelEdit,
  onSaveDisplayName,
  onDraftDisplayNameChange,
  controlsLocked = false,
}: SettingsShellProps) {
  const controlsDisabled = controlsLocked || isUpdating;

  return (
    <div className="vyn-ops-page">
      <AppliancePageHeader
        title="Settings"
        status="System settings available now · network and notifications expand later"
      />

      <AppliancePanel title="System">
        <div className="vyn-ops-settings-group">
          <div className="vyn-ops-settings-row">
            <span>Name</span>
            {editMode ? (
              <input
                aria-label="Instance name"
                value={draftDisplayName}
                onChange={(event) => onDraftDisplayNameChange(event.target.value)}
                disabled={controlsDisabled}
              />
            ) : (
              <span>{settings.instance.name}</span>
            )}
          </div>
          <div className="vyn-ops-settings-row">
            <span>Version</span>
            <span>{settings.instance.version}</span>
          </div>
          <div className="vyn-ops-settings-row">
            <span>Environment</span>
            <span>{settings.api.environment}</span>
          </div>
          <div className="vyn-ops-actions">
            {editMode ? (
              <>
                <button
                  type="button"
                  className="vyn-ops-button vyn-ops-button-primary"
                  onClick={onSaveDisplayName}
                  disabled={controlsDisabled}
                >
                  {isUpdating ? "Saving..." : "Save"}
                </button>
                <button
                  type="button"
                  className="vyn-ops-button vyn-ops-button-secondary"
                  onClick={onCancelEdit}
                  disabled={controlsDisabled}
                >
                  Cancel
                </button>
              </>
            ) : (
              <button
                type="button"
                className="vyn-ops-button vyn-ops-button-secondary"
                onClick={onStartEdit}
                disabled={controlsDisabled}
              >
                Edit name
              </button>
            )}
          </div>
          {updateValidationMessage ? <p role="alert">{updateValidationMessage}</p> : null}
          {updateErrorMessage ? <p role="alert">{updateErrorMessage}</p> : null}
        </div>
      </AppliancePanel>

      <AppliancePanel
        title="Network"
        planned
        note="Interface bonding, bridging, DNS, and time servers are not configurable in this release."
      >
        <div className="vyn-ops-settings-group vyn-ops-settings-group-disabled" aria-disabled="true">
          <div className="vyn-ops-settings-row">
            <span>Interfaces</span>
            <span>Not available yet</span>
          </div>
          <div className="vyn-ops-settings-row">
            <span>DNS / NTP</span>
            <span>Not available yet</span>
          </div>
        </div>
      </AppliancePanel>

      <AppliancePanel
        title="Notifications"
        planned
        note="Event delivery and alert preferences require a notifications backend."
      >
        <div className="vyn-ops-settings-group vyn-ops-settings-group-disabled" aria-disabled="true">
          <div className="vyn-ops-settings-row">
            <span>Web console alerts</span>
            <span>Not available yet</span>
          </div>
          <div className="vyn-ops-settings-row">
            <span>Email / webhook</span>
            <span>Not available yet</span>
          </div>
        </div>
      </AppliancePanel>
    </div>
  );
}
