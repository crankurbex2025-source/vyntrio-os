import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import type { ComponentProps } from "react";
import { describe, expect, it, vi } from "vitest";
import { StorageShell } from "./StorageShell";
import type { StorageLayoutDto } from "./storageDto";

describe("StorageShell", () => {
  const layout: StorageLayoutDto = {
    disks: {
      collected_at: "2026-07-16T10:00:00.000000000Z",
      status: "ok",
      disks: [
        {
          id: "disk-abc123",
          status: "ok",
          size_bytes: 1000000000000,
          rotational: false,
          eligibility: "eligible",
        },
        {
          id: "disk-def456",
          status: "ok",
          size_bytes: 500000000000,
          eligibility: "excluded",
          reasons: ["root_disk"],
        },
      ],
    },
    pools: {
      collected_at: "2026-07-16T10:00:00.000000000Z",
      status: "ok",
      inventory_status: "ok",
      pools: [],
      pool_management: "declared_pools",
      mutation_available: true,
      disk_format_applied: false,
      note: "Declared pools reserve eligible disks in appliance state.",
    },
    shares: {
      collected_at: "2026-07-16T10:00:00.000000000Z",
      status: "ok",
      inventory_status: "ok",
      shares: [],
      share_management: "planned_shares",
      protocol_support: "not_available",
      mutation_available: true,
    },
  };

  function renderShell(overrides: Partial<ComponentProps<typeof StorageShell>> = {}) {
    const onCreatePool = vi.fn(async () => undefined);
    const onAddDataset = vi.fn(async () => undefined);
    render(
      <StorageShell
        layout={layout}
        mutationPending={false}
        mutationError={null}
        onCreatePool={onCreatePool}
        onAddDataset={onAddDataset}
        {...overrides}
      />
    );
    return { onCreatePool };
  }

  it("renders table-first storage management", () => {
    renderShell();
    expect(screen.getByRole("heading", { name: "Storage" })).toBeInTheDocument();
    expect(screen.getByRole("table", { name: "Block device inventory" })).toBeInTheDocument();
    expect(screen.getByText("No pools declared yet.")).toBeInTheDocument();
    expect(screen.getByText("disk-abc123")).toBeInTheDocument();
    expect(screen.getByText(/Eligible candidate/i)).toBeInTheDocument();
    expect(screen.getByText(/Carries the root filesystem/i)).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Declare pool" })).toBeInTheDocument();
    expect(screen.getByText(/Disk formatting, RAID creation/i)).toBeInTheDocument();
    expect(screen.getByText(/format applied: no/i)).toBeInTheDocument();
  });

  it("declares a pool after confirm", async () => {
    const { onCreatePool } = renderShell();
    fireEvent.click(screen.getByLabelText(/disk-abc123/i));
    fireEvent.click(screen.getByLabelText(/I confirm these disks/i));
    fireEvent.click(screen.getByRole("button", { name: "Declare pool" }));
    expect(onCreatePool).toHaveBeenCalledWith("tank", ["disk-abc123"]);
  });
});
