import type { PublicSettingsDto } from "./settingsDto";

type SettingsShellProps = {
  settings: PublicSettingsDto;
};

export function SettingsShell({ settings }: SettingsShellProps) {
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
      </section>
    </main>
  );
}
