import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import type { ComponentProps } from "react";
import { describe, expect, it, vi } from "vitest";
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
  };

  function renderShell(overrides: Partial<ComponentProps<typeof OverviewShell>> = {}) {
    const onOpenSettings = vi.fn();
    const onSignOut = vi.fn();
    render(
      <OverviewShell
        overview={overview}
        isSigningOut={false}
        signOutError={false}
        settingsAccessError={false}
        settingsLoading={false}
        onOpenSettings={onOpenSettings}
        onSignOut={onSignOut}
        {...overrides}
      />
    );
    return { onOpenSettings, onSignOut };
  }

  it("renders host metric cards when metrics are available", () => {
    renderShell();

    expect(screen.getByText("Host metrics")).toBeInTheDocument();
    expect(screen.getByText("4 cores")).toBeInTheDocument();
    expect(screen.getByText(/1-minute load 0.42/)).toBeInTheDocument();
    expect(screen.getByText(/4 GB used/)).toBeInTheDocument();
    expect(screen.getByText(/State storage/)).toBeInTheDocument();
    expect(screen.getByText("No backup recorded")).toBeInTheDocument();
    expect(screen.getByText("Local network interface present")).toBeInTheDocument();
    expect(screen.getByText("0.2.0-dev")).toBeInTheDocument();
    expect(screen.getByText(/Build abc123 · development channel/)).toBeInTheDocument();
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
    expect(screen.getByText("Local network presence unclear")).toBeInTheDocument();
    expect(
      screen.getByText(/No eligible interface was observed from this process/)
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

    expect(screen.getAllByText("Unavailable").length).toBeGreaterThanOrEqual(3);
    expect(screen.getByText("CPU metrics could not be collected")).toBeInTheDocument();
    expect(screen.getByText("Memory metrics could not be collected")).toBeInTheDocument();
    expect(screen.getByText("Storage metrics could not be collected")).toBeInTheDocument();
  });

  it("invokes settings and sign out actions", () => {
    const { onOpenSettings, onSignOut } = renderShell();

    fireEvent.click(screen.getByRole("button", { name: "Instance settings" }));
    fireEvent.click(screen.getByRole("button", { name: "Sign out" }));

    expect(onOpenSettings).toHaveBeenCalledTimes(1);
    expect(onSignOut).toHaveBeenCalledTimes(1);
  });
});
