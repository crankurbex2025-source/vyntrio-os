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
  isSigningOut: boolean;
  signOutError: boolean;
  onSignOut: () => void;
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
  isSigningOut,
  signOutError,
  onSignOut,
}: SettingsShellProps) {
  const controlsDisabled = isSigningOut || isUpdating;

  return (
    <main className="settings-wrap">
      <section className="settings-card">
        <h1>Instance settings</h1>

        <div className="settings-row">
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
        <div className="settings-row">
          <span>Version</span>
          <span>{settings.instance.version}</span>
        </div>
        <div className="settings-row">
          <span>Environment</span>
          <span>{settings.api.environment}</span>
        </div>

        {editMode ? (
          <div className="settings-actions">
            <button type="button" onClick={onSaveDisplayName} disabled={controlsDisabled}>
              {isUpdating ? "Saving..." : "Save"}
            </button>
            <button type="button" onClick={onCancelEdit} disabled={controlsDisabled}>
              Cancel
            </button>
          </div>
        ) : (
          <button type="button" onClick={onStartEdit} disabled={controlsDisabled}>
            Edit name
          </button>
        )}

        {updateValidationMessage ? <p role="alert">{updateValidationMessage}</p> : null}
        {updateErrorMessage ? <p role="alert">{updateErrorMessage}</p> : null}

        <button type="button" onClick={onSignOut} disabled={controlsDisabled}>
          {isSigningOut ? "Signing out..." : "Sign out"}
        </button>
        {signOutError ? (
          <p role="alert">Sign-out could not be completed. Please try again.</p>
        ) : null}
      </section>
    </main>
  );
}
