import { useMemo, useState } from "react";
import { ApplianceDataTable } from "./ApplianceDataTable";
import { AppliancePageHeader } from "./AppliancePageHeader";
import { AppliancePanel } from "./AppliancePanel";
import "../appliance-chrome.css";
import "./appliance-demo.css";
import "./appliance-ops.css";

type DemoSection =
  | "dashboard"
  | "main"
  | "shares"
  | "users"
  | "settings"
  | "plugins"
  | "docker"
  | "vms"
  | "apps"
  | "tools";

const SECTIONS: Array<{ id: DemoSection; label: string }> = [
  { id: "dashboard", label: "Dashboard" },
  { id: "main", label: "Main" },
  { id: "shares", label: "Shares" },
  { id: "users", label: "Users" },
  { id: "settings", label: "Settings" },
  { id: "plugins", label: "Plugins" },
  { id: "docker", label: "Docker" },
  { id: "vms", label: "VMs" },
  { id: "apps", label: "Apps" },
  { id: "tools", label: "Tools" },
];

type ContainerRow = {
  id: string;
  name: string;
  image: string;
  status: "started" | "stopped";
};

type CatalogApp = {
  id: string;
  name: string;
  summary: string;
  category: string;
};

const CATALOG: CatalogApp[] = [
  {
    id: "nextcloud",
    name: "Nextcloud",
    summary: "Self-hosted files and collaboration.",
    category: "Cloud",
  },
  {
    id: "plex",
    name: "Plex Media Server",
    summary: "Organize and stream personal media.",
    category: "Media",
  },
  {
    id: "homeassistant",
    name: "Home Assistant",
    summary: "Local home automation hub.",
    category: "Automation",
  },
  {
    id: "jellyfin",
    name: "Jellyfin",
    summary: "Free software media system.",
    category: "Media",
  },
  {
    id: "grafana",
    name: "Grafana",
    summary: "Metrics dashboards and alerts.",
    category: "Observability",
  },
  {
    id: "postgres",
    name: "PostgreSQL",
    summary: "Relational database for app stacks.",
    category: "Database",
  },
];

const INITIAL_CONTAINERS: ContainerRow[] = [
  { id: "c1", name: "nextcloud", image: "nextcloud:latest", status: "started" },
  { id: "c2", name: "plex", image: "plexinc/pms-docker", status: "started" },
  { id: "c3", name: "watchtower", image: "containrrr/watchtower", status: "stopped" },
];

function UtilBar({
  pct,
  tone = "accent",
}: {
  pct: number;
  tone?: "accent" | "ok" | "warn";
}) {
  return (
    <div className={`vyn-demo-bar ${tone === "accent" ? "" : tone}`}>
      <span style={{ width: `${Math.max(0, Math.min(100, pct))}%` }} />
    </div>
  );
}

function DashboardSection({ containers }: { containers: ContainerRow[] }) {
  const started = containers.filter((c) => c.status === "started").length;
  return (
    <div className="vyn-ops-page" style={{ paddingTop: "0.85rem" }}>
      <AppliancePageHeader
        title="Dashboard"
        status="Demo widgets — host · services · storage in one viewport"
      />
      <div className="vyn-demo-widgets">
        <article className="vyn-demo-widget">
          <div className="vyn-demo-widget-head">
            <h2>System</h2>
          </div>
          <div className="vyn-demo-widget-body">
            <div className="vyn-demo-row">
              <span>Instance</span>
              <span className="vyn-demo-row-meta">Vyntrio Lab</span>
            </div>
            <div className="vyn-demo-row">
              <span>Uptime</span>
              <span className="vyn-demo-row-meta">3d 4h (demo)</span>
            </div>
            <div className="vyn-demo-row">
              <span>CPU · 8 cores</span>
              <span className="vyn-demo-row-meta">42%</span>
            </div>
            <UtilBar pct={42} tone="ok" />
            <div className="vyn-demo-row">
              <span>Memory</span>
              <span className="vyn-demo-row-meta">9.1 / 16 GB</span>
            </div>
            <UtilBar pct={57} />
            <div className="vyn-demo-row">
              <span>Board temp</span>
              <span className="vyn-demo-row-meta">48°C (demo)</span>
            </div>
          </div>
        </article>

        <article className="vyn-demo-widget">
          <div className="vyn-demo-widget-head">
            <h2>Docker</h2>
            <span className="vyn-demo-row-meta">
              {started}/{containers.length} started
            </span>
          </div>
          <div className="vyn-demo-widget-body">
            {containers.map((c) => (
              <div key={c.id} className="vyn-demo-row">
                <span>
                  <span
                    className={`vyn-demo-status-dot ${c.status === "started" ? "ok" : "stopped"}`}
                  />
                  {c.name}
                </span>
                <span className="vyn-demo-row-meta">{c.status}</span>
              </div>
            ))}
            <p className="vyn-ops-empty">Open Docker / Apps to manage or install containers.</p>
          </div>
        </article>

        <article className="vyn-demo-widget">
          <div className="vyn-demo-widget-head">
            <h2>Array / pools</h2>
          </div>
          <div className="vyn-demo-widget-body">
            <div className="vyn-demo-row">
              <span>Array</span>
              <span className="vyn-demo-row-meta">Started (demo)</span>
            </div>
            <div className="vyn-demo-row">
              <span>Capacity</span>
              <span className="vyn-demo-row-meta">12.4 / 28 TB</span>
            </div>
            <UtilBar pct={44} tone="ok" />
            <div className="vyn-demo-row">
              <span>Parity check</span>
              <span className="vyn-demo-row-meta">Idle</span>
            </div>
            <div className="vyn-demo-row">
              <span>Cache</span>
              <span className="vyn-demo-row-meta">410 / 900 GB</span>
            </div>
            <UtilBar pct={46} tone="warn" />
            <div className="vyn-demo-row">
              <span>VMs</span>
              <span className="vyn-demo-row-meta">1 running · 1 stopped</span>
            </div>
          </div>
        </article>
      </div>
    </div>
  );
}

function MainSection() {
  return (
    <div className="vyn-ops-page" style={{ paddingTop: "0.85rem" }}>
      <AppliancePageHeader
        title="Main"
        status="Array · cache · devices (demo inventory)"
      />
      <AppliancePanel title="Array devices" note="Demo rows — format/parity apply not connected.">
        <ApplianceDataTable
          caption="Array devices"
          columns={[
            { key: "device", header: "Device" },
            { key: "role", header: "Role" },
            { key: "temp", header: "Temp" },
            { key: "smart", header: "SMART" },
            { key: "used", header: "Used" },
          ]}
          rows={[
            {
              device: (
                <span>
                  <span className="vyn-demo-status-dot ok" />
                  parity
                </span>
              ),
              role: "Parity",
              temp: "39°C",
              smart: "Healthy",
              used: "—",
            },
            {
              device: (
                <span>
                  <span className="vyn-demo-status-dot ok" />
                  disk1
                </span>
              ),
              role: "Data",
              temp: "41°C",
              smart: "Healthy",
              used: (
                <div style={{ minWidth: "8rem" }}>
                  <div className="vyn-demo-row-meta">4.2 / 8 TB</div>
                  <UtilBar pct={52} tone="ok" />
                </div>
              ),
            },
            {
              device: (
                <span>
                  <span className="vyn-demo-status-dot ok" />
                  disk2
                </span>
              ),
              role: "Data",
              temp: "43°C",
              smart: "Healthy",
              used: (
                <div style={{ minWidth: "8rem" }}>
                  <div className="vyn-demo-row-meta">3.1 / 8 TB</div>
                  <UtilBar pct={39} tone="ok" />
                </div>
              ),
            },
          ]}
        />
      </AppliancePanel>
      <div className="vyn-ops-grid-2">
        <AppliancePanel title="Cache" note="Demo SSD pool.">
          <div className="vyn-demo-row">
            <span>cache</span>
            <span className="vyn-demo-row-meta">410 / 900 GB</span>
          </div>
          <UtilBar pct={46} />
        </AppliancePanel>
        <AppliancePanel title="Unassigned" note="Attach disks here in a real appliance.">
          <p className="vyn-ops-empty">No unassigned devices in this demo.</p>
        </AppliancePanel>
      </div>
      <div className="vyn-ops-actions">
        <button type="button" className="vyn-ops-button vyn-ops-button-primary" disabled>
          Start array (demo)
        </button>
        <button type="button" className="vyn-ops-button vyn-ops-button-secondary" disabled>
          Stop array (demo)
        </button>
        <button type="button" className="vyn-ops-button vyn-ops-button-secondary" disabled>
          Parity check (demo)
        </button>
      </div>
    </div>
  );
}

function DockerSection({
  containers,
  onToggle,
}: {
  containers: ContainerRow[];
  onToggle: (id: string) => void;
}) {
  return (
    <div className="vyn-ops-page" style={{ paddingTop: "0.85rem" }}>
      <AppliancePageHeader
        title="Docker"
        status="Container runtime management (demo) — install via Apps"
      />
      <AppliancePanel title="Containers">
        <ApplianceDataTable
          caption="Docker containers"
          columns={[
            { key: "name", header: "Name" },
            { key: "image", header: "Image" },
            { key: "status", header: "Status" },
            { key: "actions", header: "Actions" },
          ]}
          rows={containers.map((c) => ({
            name: (
              <span>
                <span
                  className={`vyn-demo-status-dot ${c.status === "started" ? "ok" : "stopped"}`}
                />
                {c.name}
              </span>
            ),
            image: <span className="vyn-demo-row-meta">{c.image}</span>,
            status: c.status,
            actions: (
              <button
                type="button"
                className="vyn-ops-button vyn-ops-button-secondary"
                onClick={() => onToggle(c.id)}
              >
                {c.status === "started" ? "Stop" : "Start"}
              </button>
            ),
          }))}
          emptyMessage="No containers — install from Apps."
        />
      </AppliancePanel>
    </div>
  );
}

function AppsSection({
  installed,
  onInstall,
}: {
  installed: Set<string>;
  onInstall: (app: CatalogApp) => void;
}) {
  const [selected, setSelected] = useState<CatalogApp | null>(null);
  const [path, setPath] = useState("/mnt/user/appdata");
  const [port, setPort] = useState("8080");

  return (
    <div className="vyn-ops-page" style={{ paddingTop: "0.85rem" }}>
      <AppliancePageHeader
        title="Apps"
        status="Community catalog demo — Install adds a mock Docker container"
      />
      <div className="vyn-demo-split">
        <AppliancePanel title="Catalog" note="Mock templates. No real registry pull.">
          <div className="vyn-demo-catalog">
            {CATALOG.map((app) => {
              const isInstalled = installed.has(app.id);
              return (
                <article key={app.id} className="vyn-demo-catalog-card">
                  <h3>{app.name}</h3>
                  <p>{app.summary}</p>
                  <span className="vyn-demo-row-meta">{app.category}</span>
                  <div className="vyn-ops-actions">
                    <button
                      type="button"
                      className="vyn-ops-button vyn-ops-button-primary"
                      disabled={isInstalled}
                      onClick={() => setSelected(app)}
                    >
                      {isInstalled ? "Installed" : "Install"}
                    </button>
                  </div>
                </article>
              );
            })}
          </div>
        </AppliancePanel>

        <AppliancePanel
          title="Install"
          note={
            selected
              ? `Configure ${selected.name} (demo form only).`
              : "Select an app from the catalog."
          }
        >
          {!selected ? (
            <p className="vyn-ops-empty">Choose Install on a catalog card.</p>
          ) : (
            <>
              <label className="vyn-ops-field">
                <span>App</span>
                <input value={selected.name} readOnly />
              </label>
              <label className="vyn-ops-field">
                <span>Appdata path</span>
                <input value={path} onChange={(e) => setPath(e.target.value)} />
              </label>
              <label className="vyn-ops-field">
                <span>Host port</span>
                <input value={port} onChange={(e) => setPort(e.target.value)} />
              </label>
              <div className="vyn-ops-actions">
                <button
                  type="button"
                  className="vyn-ops-button vyn-ops-button-primary"
                  onClick={() => {
                    onInstall(selected);
                    setSelected(null);
                  }}
                >
                  Apply & start (demo)
                </button>
                <button
                  type="button"
                  className="vyn-ops-button vyn-ops-button-secondary"
                  onClick={() => setSelected(null)}
                >
                  Cancel
                </button>
              </div>
            </>
          )}
        </AppliancePanel>
      </div>
    </div>
  );
}

function VmsSection() {
  return (
    <div className="vyn-ops-page" style={{ paddingTop: "0.85rem" }}>
      <AppliancePageHeader title="VMs" status="Hypervisor management (demo)" />
      <AppliancePanel title="Virtual machines">
        <ApplianceDataTable
          caption="VMs"
          columns={[
            { key: "name", header: "Name" },
            { key: "cpus", header: "vCPUs" },
            { key: "mem", header: "Memory" },
            { key: "status", header: "Status" },
          ]}
          rows={[
            {
              name: (
                <span>
                  <span className="vyn-demo-status-dot ok" />
                  win11-workstation
                </span>
              ),
              cpus: "4",
              mem: "8 GB",
              status: "started",
            },
            {
              name: (
                <span>
                  <span className="vyn-demo-status-dot stopped" />
                  debian-lab
                </span>
              ),
              cpus: "2",
              mem: "4 GB",
              status: "stopped",
            },
          ]}
        />
      </AppliancePanel>
      <div className="vyn-ops-actions">
        <button type="button" className="vyn-ops-button vyn-ops-button-primary" disabled>
          Add VM (demo)
        </button>
      </div>
    </div>
  );
}

function SharesSection() {
  return (
    <div className="vyn-ops-page" style={{ paddingTop: "0.85rem" }}>
      <AppliancePageHeader title="Shares" status="User + disk shares (demo)" />
      <AppliancePanel title="User shares">
        <ApplianceDataTable
          caption="User shares"
          columns={[
            { key: "name", header: "Name" },
            { key: "useCache", header: "Cache" },
            { key: "export", header: "SMB" },
            { key: "security", header: "Security" },
          ]}
          rows={[
            { name: "media", useCache: "Prefer", export: "Public (demo)", security: "Public" },
            { name: "appdata", useCache: "Prefer", export: "Private", security: "Private" },
            { name: "domains", useCache: "No", export: "Private", security: "Private" },
          ]}
        />
      </AppliancePanel>
      <div className="vyn-ops-actions">
        <button type="button" className="vyn-ops-button vyn-ops-button-primary" disabled>
          Add share (demo)
        </button>
      </div>
    </div>
  );
}

function UsersSection() {
  return (
    <div className="vyn-ops-page" style={{ paddingTop: "0.85rem" }}>
      <AppliancePageHeader title="Users" status="Local accounts (demo)" />
      <AppliancePanel title="Accounts">
        <ApplianceDataTable
          caption="Users"
          columns={[
            { key: "user", header: "User" },
            { key: "desc", header: "Description" },
            { key: "role", header: "Role" },
          ]}
          rows={[
            { user: "root", desc: "Administrator", role: "Owner" },
            { user: "media", desc: "SMB media access", role: "Share user" },
          ]}
        />
      </AppliancePanel>
    </div>
  );
}

function SettingsSection() {
  return (
    <div className="vyn-ops-page" style={{ paddingTop: "0.85rem" }}>
      <AppliancePageHeader title="Settings" status="System · network · identification (demo groups)" />
      <div className="vyn-ops-grid-2">
        <AppliancePanel title="Identification">
          <div className="vyn-ops-settings-row">
            <span>Server name</span>
            <span>Vyntrio Lab</span>
          </div>
          <div className="vyn-ops-settings-row">
            <span>Description</span>
            <span>Demo appliance</span>
          </div>
        </AppliancePanel>
        <AppliancePanel title="Network">
          <div className="vyn-ops-settings-row">
            <span>IPv4</span>
            <span className="vyn-demo-row-meta">10.0.0.50/24 (demo)</span>
          </div>
          <div className="vyn-ops-settings-row">
            <span>DNS</span>
            <span className="vyn-demo-row-meta">1.1.1.1 (demo)</span>
          </div>
        </AppliancePanel>
        <AppliancePanel title="Docker">
          <div className="vyn-ops-settings-row">
            <span>Enable Docker</span>
            <span>Yes (demo)</span>
          </div>
          <div className="vyn-ops-settings-row">
            <span>vDisk path</span>
            <span className="vyn-demo-row-meta">/mnt/user/system/docker.img</span>
          </div>
        </AppliancePanel>
        <AppliancePanel title="VM Manager">
          <div className="vyn-ops-settings-row">
            <span>Enable VMs</span>
            <span>Yes (demo)</span>
          </div>
          <div className="vyn-ops-settings-row">
            <span>Libvirt path</span>
            <span className="vyn-demo-row-meta">/mnt/user/system/libvirt.img</span>
          </div>
        </AppliancePanel>
      </div>
    </div>
  );
}

function PluginsSection() {
  return (
    <div className="vyn-ops-page" style={{ paddingTop: "0.85rem" }}>
      <AppliancePageHeader title="Plugins" status="Extensions (demo)" />
      <AppliancePanel title="Installed plugins">
        <ApplianceDataTable
          caption="Plugins"
          columns={[
            { key: "name", header: "Plugin" },
            { key: "version", header: "Version" },
            { key: "status", header: "Status" },
          ]}
          rows={[
            { name: "Community Applications (demo)", version: "2026.07", status: "Active" },
            { name: "Custom CSS (demo)", version: "0.2.0", status: "Active" },
          ]}
        />
      </AppliancePanel>
      <div className="vyn-ops-actions">
        <button type="button" className="vyn-ops-button vyn-ops-button-secondary" disabled>
          Install plugin (demo)
        </button>
      </div>
    </div>
  );
}

function ToolsSection() {
  return (
    <div className="vyn-ops-page" style={{ paddingTop: "0.85rem" }}>
      <AppliancePageHeader title="Tools" status="Logs · diagnostics · notifications (demo)" />
      <div className="vyn-ops-grid-2">
        <AppliancePanel title="System log" note="Tail preview (demo lines).">
          <pre
            className="vyn-demo-row-meta"
            style={{ margin: 0, whiteSpace: "pre-wrap", lineHeight: 1.45 }}
          >
            {`Jul 18 19:00:01 vyntrio kernel: demo boot complete
Jul 18 19:00:12 vyntrio docker: nextcloud started
Jul 18 19:01:04 vyntrio array: parity idle
Jul 18 19:02:22 vyntrio smb: share media exported (demo)`}
          </pre>
        </AppliancePanel>
        <AppliancePanel title="Notifications">
          <div className="vyn-demo-row">
            <span>
              <span className="vyn-demo-status-dot ok" />
              Array started
            </span>
            <span className="vyn-demo-row-meta">2h ago</span>
          </div>
          <div className="vyn-demo-row">
            <span>
              <span className="vyn-demo-status-dot warn" />
              Cache above 45%
            </span>
            <span className="vyn-demo-row-meta">18m ago</span>
          </div>
        </AppliancePanel>
        <AppliancePanel title="Diagnostics">
          <button type="button" className="vyn-ops-button vyn-ops-button-secondary" disabled>
            Download diagnostics zip (demo)
          </button>
        </AppliancePanel>
        <AppliancePanel title="Scheduled tasks">
          <div className="vyn-demo-row">
            <span>Parity check</span>
            <span className="vyn-demo-row-meta">Monthly · 1st 00:00</span>
          </div>
          <div className="vyn-demo-row">
            <span>Mover</span>
            <span className="vyn-demo-row-meta">Daily · 03:40</span>
          </div>
        </AppliancePanel>
      </div>
    </div>
  );
}

/**
 * Full Unraid-class WebGUI demo for visual QA.
 * Mock data only — not production management.
 */
export function ApplianceFullDemo() {
  const [section, setSection] = useState<DemoSection>("dashboard");
  const [containers, setContainers] = useState<ContainerRow[]>(INITIAL_CONTAINERS);

  const installed = useMemo(() => new Set(containers.map((c) => c.name)), [containers]);

  function toggleContainer(id: string) {
    setContainers((prev) =>
      prev.map((c) =>
        c.id === id
          ? { ...c, status: c.status === "started" ? "stopped" : "started" }
          : c
      )
    );
  }

  function installApp(app: CatalogApp) {
    setContainers((prev) => {
      if (prev.some((c) => c.name === app.id)) {
        return prev;
      }
      return [
        ...prev,
        {
          id: `c-${app.id}`,
          name: app.id,
          image: `${app.id}:latest`,
          status: "started",
        },
      ];
    });
    setSection("docker");
  }

  const startedCount = containers.filter((c) => c.status === "started").length;

  return (
    <div className="vyn-appliance-shell vyn-appliance-shell-ops" style={{ minHeight: "100vh" }}>
      <div className="vyn-demo-banner" role="status">
        <div>
          <strong>TEST DEMO</strong> — Full Unraid-class WebGUI surface with mock data. Docker
          Install is interactive in-browser only · no real runtime · StyleSeed{" "}
          <code>operations-console</code>.
        </div>
        <a href="/app" style={{ color: "var(--vyn-accent)" }}>
          Open real /app
        </a>
      </div>

      <nav className="vyn-demo-topnav" aria-label="Demo sections">
        {SECTIONS.map((item) => (
          <button
            key={item.id}
            type="button"
            className={section === item.id ? "active" : undefined}
            onClick={() => setSection(item.id)}
          >
            {item.label}
          </button>
        ))}
      </nav>

      {section === "dashboard" ? <DashboardSection containers={containers} /> : null}
      {section === "main" ? <MainSection /> : null}
      {section === "shares" ? <SharesSection /> : null}
      {section === "users" ? <UsersSection /> : null}
      {section === "settings" ? <SettingsSection /> : null}
      {section === "plugins" ? <PluginsSection /> : null}
      {section === "docker" ? (
        <DockerSection containers={containers} onToggle={toggleContainer} />
      ) : null}
      {section === "vms" ? <VmsSection /> : null}
      {section === "apps" ? (
        <AppsSection installed={installed} onInstall={installApp} />
      ) : null}
      {section === "tools" ? <ToolsSection /> : null}

      <footer className="vyn-demo-footer">
        <span>
          Array: <span className="ok">Started (demo)</span>
        </span>
        <span>
          Docker:{" "}
          <span className="ok">
            {startedCount}/{containers.length} started
          </span>
        </span>
        <span>
          Parity: <span className="warn">Idle</span>
        </span>
        <span>Vyntrio demo · not Unraid</span>
      </footer>
    </div>
  );
}
