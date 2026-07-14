import { describe, expect, it } from "vitest";
import { formatMetricBytes, parseOverviewDto } from "./overviewDto";

describe("parseOverviewDto", () => {
  function validHost() {
    return {
      cpu: {
        status: "ok" as const,
        logical_cores: 4,
        load_1m: 0.42,
      },
      memory: {
        status: "ok" as const,
        total_bytes: 8589934592,
        available_bytes: 4294967296,
        used_bytes: 4294967296,
      },
      filesystems: [
        {
          id: "state" as const,
          status: "ok" as const,
          total_bytes: 107374182400,
          available_bytes: 53687091200,
          used_bytes: 53687091200,
          fs_type: "ext4" as const,
        },
      ],
    };
  }

  function validBackup() {
    return {
      status: "never_run" as const,
      ever_succeeded: false,
    };
  }

  function validNetwork() {
    return {
      status: "available" as const,
    };
  }

  function validPayload() {
    return {
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
      host: validHost(),
      backup: validBackup(),
      network: validNetwork(),
      collected_at: "2026-07-14T12:00:00.000000000Z",
    };
  }

  it("accepts the canonical overview payload with host metrics", () => {
    expect(parseOverviewDto(validPayload())).toEqual(validPayload());
  });

  it("accepts unavailable host metric sections without numeric fields", () => {
    const payload = {
      ...validPayload(),
      host: {
        cpu: { status: "unavailable" },
        memory: { status: "unavailable" },
        filesystems: [{ id: "state", status: "unavailable" }],
      },
    };
    expect(parseOverviewDto(payload)).toEqual(payload);
  });

  it("rejects host memory with mismatched used bytes", () => {
    expect(
      parseOverviewDto({
        ...validPayload(),
        host: {
          ...validHost(),
          memory: {
            status: "ok",
            total_bytes: 100,
            available_bytes: 40,
            used_bytes: 50,
          },
        },
      })
    ).toBeNull();
  });

  it("accepts unavailable backup status without extra fields", () => {
    expect(parseOverviewDto({ ...validPayload(), backup: { status: "unavailable" } })).toEqual({
      ...validPayload(),
      backup: { status: "unavailable" },
    });
  });

  it("rejects failed backup without failure class", () => {
    expect(
      parseOverviewDto({
        ...validPayload(),
        backup: {
          status: "failed",
          completed_at: "2026-07-14T11:30:00.000000000Z",
          ever_succeeded: false,
        },
      })
    ).toBeNull();
  });

  it("rejects unknown top-level fields", () => {
    expect(parseOverviewDto({ ...validPayload(), extra: "x" })).toBeNull();
  });

  it("accepts all valid network status values", () => {
    for (const status of ["available", "unknown", "unavailable"] as const) {
      expect(parseOverviewDto({ ...validPayload(), network: { status } })).toEqual({
        ...validPayload(),
        network: { status },
      });
    }
  });

  it("rejects network with invalid status", () => {
    expect(parseOverviewDto({ ...validPayload(), network: { status: "connected" } })).toBeNull();
  });

  it("rejects network with missing status", () => {
    expect(parseOverviewDto({ ...validPayload(), network: {} })).toBeNull();
  });

  it("rejects network with extra fields", () => {
    expect(
      parseOverviewDto({ ...validPayload(), network: { status: "available", interface: "eth0" } })
    ).toBeNull();
  });

  it("formats metric bytes for display", () => {
    expect(formatMetricBytes(1024)).toBe("1 KB");
  });
});
