import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import type { ComponentProps } from "react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { OverviewShell } from "./OverviewShell";
import type { OverviewDto } from "./overviewDto";

describe("OverviewShell", () => {
  const overview: OverviewDto = {
    instance: {
      name: "Vyntrio Home",
      version: "0.2.0-dev",
      commit: "abc123",
    },
    api: {
      environment: "development",
    },
    service: {
      status: "running",
    },
    readiness: {
      status: "ready",
      database: "ok",
    },
    host: {
      cpu: {
        status: "ok",
        logical_cores: 4,
        load_1m: 0.42,
      },
      memory: {
        status: "ok",
        total_bytes: 8589934592,
        available_bytes: 4294967296,
        used_bytes: 4294967296,
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
    collected_at: "2026-07-14T12:00:00.000000000Z",
    backup: {
      status: "never_run",
      ever_succeeded: false,
    },
    network: {
      status: "available",
    },
    software: {
      status: "ok",
      version: "0.2.0-dev",
      commit: "abc123",
      channel: "development",
    },
    runtime: {
      status: "ready",
    },
    health: {
      status: "healthy",
    },
    storage: {
      status: "ok",
      disk_count: 2,
      eligible_count: 1,
      excluded_count: 1,
      unknown_count: 0,
      pool_count: 0,
      share_count: 0,
      mutation_available: true,
    },
  };

  function renderShell(overrides: Partial<ComponentProps<typeof OverviewShell>> = {}) {
    render(
      <MemoryRouter>
        <OverviewShell
          overview={overview}
          signOutError={false}
          settingsAccessError={false}
          storageAccessError={false}
          settingsLoading={false}
          storageLoading={false}
          {...overrides}
        />
      </MemoryRouter>
    );
  }

  it("renders operational metric strip and panels", () => {
    renderShell();

    expect(screen.getByRole("heading", { name: "Dashboard" })).toBeInTheDocument();
    expect(screen.getByLabelText("System snapshot")).toBeInTheDocument();
    expect(screen.getByLabelText("Setup progress")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Host" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Storage layout" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Software" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Health and backup" })).toBeInTheDocument();
    expect(screen.getByText(/4 cores · load 0\.42/)).toBeInTheDocument();
    expect(screen.getByText(/Never run/)).toBeInTheDocument();
    expect(screen.getByText(/Interface present/)).toBeInTheDocument();
    expect(screen.getByText("0.2.0-dev")).toBeInTheDocument();
    expect(screen.getByText(/Build abc123 · development channel/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Review disks / declare pool" })).toHaveAttribute(
      "href",
      "/app/storage"
    );
    expect(screen.getByRole("link", { name: "Prepare share plan" })).toHaveAttribute(
      "href",
      "/app/shares"
    );
    expect(screen.getByRole("link", { name: "Instance name" })).toHaveAttribute(
      "href",
      "/app/settings"
    );
    expect(screen.queryByText("csrf_token")).not.toBeInTheDocument();
    expect(screen.queryByText(/eth0|192\.168|mac/i)).not.toBeInTheDocument();
  });

  it("renders unknown network presence copy", () => {
    renderShell({
      overview: {
        ...overview,
        network: { status: "unknown" },
      },
    });
    expect(screen.getByText("Unclear")).toBeInTheDocument();
    expect(
      screen.getByText(/No eligible interface observed from this process/)
    ).toBeInTheDocument();
  });

  it("renders unavailable network presence copy", () => {
    renderShell({
      overview: {
        ...overview,
        network: { status: "unavailable" },
      },
    });
    expect(screen.getByText("Network presence could not be determined.")).toBeInTheDocument();
  });

  it("renders warning health summary for backup note", () => {
    renderShell({
      overview: {
        ...overview,
        health: { status: "warning", note: "backup" },
      },
    });
    expect(screen.getByText(/last recorded local backup attempt failed/)).toBeInTheDocument();
  });

  it("renders degraded runtime readiness copy", () => {
    renderShell({
      overview: {
        ...overview,
        runtime: { status: "degraded", note: "database" },
        readiness: { status: "not_ready", database: "error" },
      },
    });
    expect(screen.getAllByText("Degraded").length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText(/Database not ready/)).toBeInTheDocument();
  });

  it("renders unavailable software release metadata", () => {
    renderShell({
      overview: {
        ...overview,
        software: { status: "unavailable" },
      },
    });
    expect(screen.getByText("Software release metadata could not be determined.")).toBeInTheDocument();
  });

  it("renders unavailable backup status explicitly", () => {
    renderShell({
      overview: {
        ...overview,
        backup: { status: "unavailable" },
      },
    });
    expect(screen.getByText("Backup status could not be read.")).toBeInTheDocument();
  });

  it("renders unavailable host metrics explicitly", () => {
    renderShell({
      overview: {
        ...overview,
        host: {
          cpu: { status: "unavailable" },
          memory: { status: "unavailable" },
          filesystems: [{ id: "state", status: "unavailable" }],
        },
      },
    });

    expect(screen.getByText("CPU metrics could not be collected")).toBeInTheDocument();
    expect(screen.getByText("Memory metrics could not be collected")).toBeInTheDocument();
    expect(screen.getByText("Storage metrics could not be collected")).toBeInTheDocument();
  });
});
