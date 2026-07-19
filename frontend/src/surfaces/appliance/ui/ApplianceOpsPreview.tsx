import { useSearchParams } from "react-router-dom";
import { OverviewShell } from "../../../features/overview/OverviewShell";
import type { OverviewDto } from "../../../features/overview/overviewDto";
import { StorageShell } from "../../../features/storage/StorageShell";
import type { StorageLayoutDto } from "../../../features/storage/storageDto";
import { SharesShell } from "../../../features/shares/SharesShell";
import { SettingsShell } from "../../../features/settings/SettingsShell";
import { UsersShell } from "../../../features/users/UsersShell";
import { ToolsShell } from "../../../features/tools/ToolsShell";
import { ApplianceShell } from "../ApplianceShell";

const overviewFixture: OverviewDto = {
  instance: { name: "Vyntrio Lab", version: "0.2.0-dev", commit: "abc123" },
  api: { environment: "development" },
  service: { status: "running" },
  readiness: { status: "ready", database: "ok" },
  host: {
    cpu: { status: "ok", logical_cores: 8, load_1m: 0.64 },
    memory: {
      status: "ok",
      total_bytes: 17179869184,
      available_bytes: 8589934592,
      used_bytes: 8589934592,
    },
    filesystems: [
      {
        id: "state",
        status: "ok",
        total_bytes: 107374182400,
        available_bytes: 53687091200,
        used_bytes: 53687091200,
        fs_type: "ext4",
      },
    ],
  },
  collected_at: "2026-07-18T12:00:00.000000000Z",
  backup: { status: "never_run", ever_succeeded: false },
  network: { status: "available" },
  software: {
    status: "ok",
    version: "0.2.0-dev",
    commit: "abc123",
    channel: "development",
  },
  runtime: { status: "ready" },
  health: { status: "healthy" },
  storage: {
    status: "ok",
    disk_count: 2,
    eligible_count: 1,
    excluded_count: 1,
    unknown_count: 0,
    pool_count: 1,
    share_count: 1,
    mutation_available: true,
  },
};

const storageFixture: StorageLayoutDto = {
  disks: {
    collected_at: "2026-07-18T12:00:00.000000000Z",
    status: "ok",
    disks: [
      {
        id: "disk-nvme0",
        status: "ok",
        size_bytes: 1000000000000,
        rotational: false,
        eligibility: "eligible",
      },
      {
        id: "disk-sda",
        status: "ok",
        size_bytes: 500000000000,
        eligibility: "excluded",
        reasons: ["root_disk"],
      },
    ],
  },
  pools: {
    collected_at: "2026-07-18T12:00:00.000000000Z",
    status: "ok",
    inventory_status: "ok",
    pools: [
      {
        id: "pool-1",
        name: "tank",
        status: "declared",
        disk_ids: ["disk-nvme0"],
        disk_format_state: "not_applied",
        created_at: "2026-07-18T11:00:00.000000000Z",
        updated_at: "2026-07-18T11:00:00.000000000Z",
        datasets: [
          {
            id: "ds-1",
            name: "data",
            path_intent: "/tank/data",
            status: "planned",
            created_at: "2026-07-18T11:00:00.000000000Z",
          },
        ],
      },
    ],
    pool_management: "declared_pools",
    mutation_available: true,
    disk_format_applied: false,
    note: "Declared pools reserve eligible disks in appliance state.",
  },
  shares: {
    collected_at: "2026-07-18T12:00:00.000000000Z",
    status: "ok",
    inventory_status: "ok",
    shares: [
      {
        id: "share-1",
        name: "media",
        pool_id: "pool-1",
        dataset_id: "ds-1",
        protocol: "planned",
        status: "planned",
        created_at: "2026-07-18T11:30:00.000000000Z",
      },
    ],
    share_management: "planned_shares",
    protocol_support: "not_available",
    mutation_available: true,
  },
};

/**
 * Auth-free StyleSeed / visual QA harness for the appliance WebGUI.
 * Not a production management surface.
 */
export function ApplianceOpsPreview() {
  const [params] = useSearchParams();
  const section = params.get("section") ?? "dashboard";
  const activeId =
    section === "storage"
      ? "storage"
      : section === "shares"
        ? "shares"
        : section === "settings"
          ? "settings"
          : section === "users"
            ? "users"
            : section === "tools"
              ? "tools"
              : "dashboard";

  return (
    <ApplianceShell withNav instanceName="Vyntrio Lab" forceActiveId={activeId}>
      {section === "storage" ? (
        <StorageShell
          layout={storageFixture}
          mutationPending={false}
          mutationError={null}
          onCreatePool={async () => undefined}
          onAddDataset={async () => undefined}
        />
      ) : null}
      {section === "shares" ? (
        <SharesShell
          layout={storageFixture}
          mutationPending={false}
          mutationError={null}
          onCreateShare={async () => undefined}
        />
      ) : null}
      {section === "settings" ? (
        <SettingsShell
          settings={{
            instance: { name: "Vyntrio Lab", version: "0.2.0-dev" },
            api: { environment: "development" },
          }}
          editMode={false}
          draftDisplayName="Vyntrio Lab"
          isUpdating={false}
          updateErrorMessage={null}
          updateValidationMessage={null}
          onStartEdit={() => undefined}
          onCancelEdit={() => undefined}
          onSaveDisplayName={() => undefined}
          onDraftDisplayNameChange={() => undefined}
        />
      ) : null}
      {section === "users" ? <UsersShell /> : null}
      {section === "tools" ? <ToolsShell /> : null}
      {section === "dashboard" ||
      !["storage", "shares", "settings", "users", "tools"].includes(section) ? (
        <OverviewShell
          overview={overviewFixture}
          signOutError={false}
          settingsAccessError={false}
          storageAccessError={false}
          settingsLoading={false}
          storageLoading={false}
        />
      ) : null}
    </ApplianceShell>
  );
}
