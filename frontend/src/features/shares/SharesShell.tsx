import { useState } from "react";
import { Link } from "react-router-dom";
import { ApplianceDataTable } from "../../surfaces/appliance/ui/ApplianceDataTable";
import { AppliancePageHeader } from "../../surfaces/appliance/ui/AppliancePageHeader";
import { AppliancePanel } from "../../surfaces/appliance/ui/AppliancePanel";
import type { StorageLayoutDto } from "../storage/storageDto";

type SharesShellProps = {
  layout: StorageLayoutDto;
  mutationPending: boolean;
  mutationError: string | null;
  onCreateShare: (name: string, poolId: string, datasetId?: string) => Promise<void>;
};

export function SharesShell({
  layout,
  mutationPending,
  mutationError,
  onCreateShare,
}: SharesShellProps) {
  const { pools, shares } = layout;
  const [shareName, setShareName] = useState("data");
  const [sharePoolId, setSharePoolId] = useState("");
  const [shareDatasetId, setShareDatasetId] = useState("");

  async function handleCreateShare() {
    const poolId = sharePoolId || pools.pools[0]?.id;
    if (!poolId || !shareName.trim()) {
      return;
    }
    await onCreateShare(shareName.trim(), poolId, shareDatasetId || undefined);
  }

  const shareRows = shares.shares.map((share) => ({
    name: share.name,
    pool: share.pool_id,
    dataset: share.dataset_id || "—",
    protocol: "Planned / not published",
    status: share.status,
  }));

  return (
    <div className="vyn-ops-page">
      <AppliancePageHeader
        title="Shares"
        status="Share plans only — SMB/NFS/iSCSI publish is not available yet"
      />

      {mutationError ? (
        <section className="vyn-ops-alert" role="alert">
          {mutationError}
        </section>
      ) : null}

      <AppliancePanel
        title="Share plans"
        note={
          shares.note ??
          "Protocol services are not started. Declared share plans reserve intent in appliance state only."
        }
      >
        <ApplianceDataTable
          caption="Share plans"
          columns={[
            { key: "name", header: "Name" },
            { key: "pool", header: "Pool" },
            { key: "dataset", header: "Dataset" },
            { key: "protocol", header: "Protocol" },
            { key: "status", header: "Status" },
          ]}
          rows={shareRows}
          emptyMessage="No share plans yet."
        />
      </AppliancePanel>

      <AppliancePanel
        title="Prepare share plan"
        note="Stores a share plan only. SMB/NFS services are not started."
      >
        {pools.pools.length === 0 ? (
          <p className="vyn-ops-empty">
            Declare a pool under <Link to="/app/storage">Storage</Link> before preparing a share
            plan.
          </p>
        ) : (
          <>
            <label className="vyn-ops-field">
              <span>Share name</span>
              <input
                value={shareName}
                onChange={(event) => setShareName(event.target.value)}
                disabled={mutationPending}
                maxLength={32}
              />
            </label>
            <label className="vyn-ops-field">
              <span>Pool</span>
              <select
                value={sharePoolId || pools.pools[0].id}
                onChange={(event) => setSharePoolId(event.target.value)}
                disabled={mutationPending}
              >
                {pools.pools.map((pool) => (
                  <option key={pool.id} value={pool.id}>
                    {pool.name}
                  </option>
                ))}
              </select>
            </label>
            <label className="vyn-ops-field">
              <span>Dataset (optional)</span>
              <select
                value={shareDatasetId}
                onChange={(event) => setShareDatasetId(event.target.value)}
                disabled={mutationPending}
              >
                <option value="">None</option>
                {(
                  pools.pools.find((pool) => pool.id === (sharePoolId || pools.pools[0].id))
                    ?.datasets ?? []
                ).map((dataset) => (
                  <option key={dataset.id} value={dataset.id}>
                    {dataset.name}
                  </option>
                ))}
              </select>
            </label>
            <div className="vyn-ops-actions">
              <button
                type="button"
                className="vyn-ops-button vyn-ops-button-secondary"
                disabled={mutationPending || !shareName.trim()}
                onClick={() => void handleCreateShare()}
              >
                {mutationPending ? "Working..." : "Prepare share plan"}
              </button>
            </div>
          </>
        )}
      </AppliancePanel>
    </div>
  );
}
