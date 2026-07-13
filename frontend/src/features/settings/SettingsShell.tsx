import type { PublicSettingsDto } from "./settingsDto";

type SettingsShellProps = {
  settings: PublicSettingsDto;
  isSigningOut: boolean;
  signOutError: boolean;
  onSignOut: () => void;
};

export function SettingsShell({ settings, isSigningOut, signOutError, onSignOut }: SettingsShellProps) {
  return (
    <main className="settings-wrap">
      <section className="settings-card">
        <h1>Instance settings</h1>

        <div className="settings-row">
          <span>Name</span>
          <span>{settings.instance.name}</span>
        </div>
        <div className="settings-row">
          <span>Version</span>
          <span>{settings.instance.version}</span>
        </div>
        <div className="settings-row">
          <span>Environment</span>
          <span>{settings.api.environment}</span>
        </div>

        <button type="button" onClick={onSignOut} disabled={isSigningOut}>
          {isSigningOut ? "Signing out..." : "Sign out"}
        </button>
        {signOutError ? (
          <p role="alert">Sign-out could not be completed. Please try again.</p>
        ) : null}
      </section>
    </main>
  );
}
