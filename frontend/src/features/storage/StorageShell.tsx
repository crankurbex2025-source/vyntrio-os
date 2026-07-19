import { useMemo, useState } from "react";
import { ApplianceDataTable } from "../../surfaces/appliance/ui/ApplianceDataTable";
import { AppliancePageHeader } from "../../surfaces/appliance/ui/AppliancePageHeader";
import { AppliancePanel } from "../../surfaces/appliance/ui/AppliancePanel";
import type { StorageLayoutDto } from "./storageDto";
import {
  eligibleDisks,
  formatDiskCollectedAt,
  formatDiskSize,
  formatEligibilityLabel,
  formatExclusionReason,
} from "./storageDto";

type StorageShellProps = {
  layout: StorageLayoutDto;
  mutationPending: boolean;
  mutationError: string | null;
  onCreatePool: (name: string, diskIds: string[]) => Promise<void>;
  onAddDataset: (poolId: string, name: string) => Promise<void>;
};

export function StorageShell({
  layout,
  mutationPending,
  mutationError,
  onCreatePool,
  onAddDataset,
}: StorageShellProps) {
  const { disks: inventory, pools } = layout;
  const eligible = useMemo(() => eligibleDisks(inventory.disks), [inventory.disks]);
  const usedDiskIds = useMemo(() => {
    const used = new Set<string>();
    for (const pool of pools.pools) {
      for (const id of pool.disk_ids) {
        used.add(id);
      }
    }
    return used;
  }, [pools.pools]);
  const availableEligible = eligible.filter((disk) => !usedDiskIds.has(disk.id));

  const [poolName, setPoolName] = useState("tank");
  const [selectedDisks, setSelectedDisks] = useState<string[]>([]);
  const [confirmPool, setConfirmPool] = useState(false);
  const [datasetName, setDatasetName] = useState("data");
  const [datasetPoolId, setDatasetPoolId] = useState("");

  const busy = mutationPending;

  function toggleDisk(id: string) {
    setSelectedDisks((previous) =>
      previous.includes(id) ? previous.filter((item) => item !== id) : [...previous, id]
    );
  }

  async function handleCreatePool() {
    if (!confirmPool || selectedDisks.length === 0 || !poolName.trim()) {
      return;
    }
    await onCreatePool(poolName.trim(), selectedDisks);
    setSelectedDisks([]);
    setConfirmPool(false);
  }

  async function handleAddDataset() {
    const poolId = datasetPoolId || pools.pools[0]?.id;
    if (!poolId || !datasetName.trim()) {
      return;
    }
    await onAddDataset(poolId, datasetName.trim());
  }

  const diskRows = inventory.disks.map((disk) => ({
    id: disk.id,
    size:
      disk.status === "unavailable" ? "Unavailable" : formatDiskSize(disk.size_bytes),
    eligibility: formatEligibilityLabel(disk.eligibility),
    assignment: usedDiskIds.has(disk.id) ? "Reserved by pool" : "Free",
    reasons:
      disk.reasons && disk.reasons.length > 0
        ? disk.reasons.map((reason) => formatExclusionReason(reason)).join("; ")
        : "—",
  }));

  const poolRows = pools.pools.map((pool) => ({
    name: pool.name,
    status: pool.status,
    disks: pool.disk_ids.join(", ") || "—",
    format: pool.disk_format_state,
    datasets:
      pool.datasets.length === 0
        ? "None"
        : pool.datasets
            .map((dataset) => `${dataset.name} (${dataset.status})`)
            .join(", "),
  }));

  return (
    <div className="vyn-ops-page">
      <AppliancePageHeader
        title="Storage"
        status={`Disk inventory and declared pools · format applied: ${pools.disk_format_applied ? "yes" : "no"} · collected ${formatDiskCollectedAt(inventory.collected_at)}`}
      />

      {mutationError ? (
        <section className="vyn-ops-alert" role="alert">
          {mutationError}
        </section>
      ) : null}

      <AppliancePanel
        title="Block devices"
        note={
          inventory.status === "ok"
            ? `${inventory.disks.length} device(s) · ${availableEligible.length} free eligible for new pools`
            : "Block device inventory could not be collected on this host."
        }
      >
        <ApplianceDataTable
          caption="Block device inventory"
          columns={[
            { key: "id", header: "Device" },
            { key: "size", header: "Size" },
            { key: "eligibility", header: "Eligibility" },
            { key: "assignment", header: "Assignment" },
            { key: "reasons", header: "Notes" },
          ]}
          rows={diskRows}
          emptyMessage="No devices reported."
        />
      </AppliancePanel>

      <AppliancePanel
        title="Declared pools"
        note={
          pools.note ??
          "Declared pools reserve disks in appliance state. Formatting and RAID apply are not available."
        }
      >
        <ApplianceDataTable
          caption="Declared pools"
          columns={[
            { key: "name", header: "Pool" },
            { key: "status", header: "Status" },
            { key: "disks", header: "Disks" },
            { key: "format", header: "Format state" },
            { key: "datasets", header: "Datasets" },
          ]}
          rows={poolRows}
          emptyMessage="No pools declared yet."
        />
      </AppliancePanel>

      <div className="vyn-ops-grid-2">
        <AppliancePanel
          title="Declare pool"
          note="Reserves disks in state only — does not wipe or format them."
        >
          {availableEligible.length === 0 ? (
            <p className="vyn-ops-empty">
              No free eligible disks. Attach unused disks or free disks from existing declared pools.
            </p>
          ) : (
            <>
              <label className="vyn-ops-field">
                <span>Pool name</span>
                <input
                  value={poolName}
                  onChange={(event) => setPoolName(event.target.value)}
                  disabled={busy}
                  maxLength={32}
                />
              </label>
              <div className="vyn-ops-check-list">
                {availableEligible.map((disk) => (
                  <label key={disk.id}>
                    <input
                      type="checkbox"
                      checked={selectedDisks.includes(disk.id)}
                      onChange={() => toggleDisk(disk.id)}
                      disabled={busy}
                      aria-label={disk.id}
                    />
                    <span>
                      {disk.id} · {formatDiskSize(disk.size_bytes)}
                    </span>
                  </label>
                ))}
              </div>
              <label className="vyn-ops-field">
                <span>
                  <input
                    type="checkbox"
                    checked={confirmPool}
                    onChange={(event) => setConfirmPool(event.target.checked)}
                    disabled={busy}
                    aria-label="I confirm these disks should be reserved for this pool (no format yet)"
                  />{" "}
                  I confirm these disks should be reserved for this pool (no format yet)
                </span>
              </label>
              <div className="vyn-ops-actions">
                <button
                  type="button"
                  className="vyn-ops-button vyn-ops-button-primary"
                  disabled={busy || !confirmPool || selectedDisks.length === 0 || !poolName.trim()}
                  onClick={() => void handleCreatePool()}
                >
                  {mutationPending ? "Working..." : "Declare pool"}
                </button>
              </div>
            </>
          )}
        </AppliancePanel>

        <AppliancePanel title="Prepare dataset" note="Dataset plans only — no filesystem created yet.">
          {pools.pools.length === 0 ? (
            <p className="vyn-ops-empty">Declare a pool first.</p>
          ) : (
            <>
              <label className="vyn-ops-field">
                <span>Pool</span>
                <select
                  value={datasetPoolId || pools.pools[0].id}
                  onChange={(event) => setDatasetPoolId(event.target.value)}
                  disabled={busy}
                >
                  {pools.pools.map((pool) => (
                    <option key={pool.id} value={pool.id}>
                      {pool.name}
                    </option>
                  ))}
                </select>
              </label>
              <label className="vyn-ops-field">
                <span>Dataset name</span>
                <input
                  value={datasetName}
                  onChange={(event) => setDatasetName(event.target.value)}
                  disabled={busy}
                  maxLength={32}
                />
              </label>
              <div className="vyn-ops-actions">
                <button
                  type="button"
                  className="vyn-ops-button vyn-ops-button-secondary"
                  disabled={busy || !datasetName.trim()}
                  onClick={() => void handleAddDataset()}
                >
                  Prepare dataset
                </button>
              </div>
            </>
          )}
        </AppliancePanel>
      </div>

      <section className="vyn-ops-alert vyn-ops-alert-status" role="status">
        Disk formatting, RAID creation, and live SMB/NFS publishing are not available. Declared
        pools and plans are real persisted appliance state.
      </section>
    </div>
  );
}
