import { describe, expect, it } from "vitest";
import { parseOverviewDto } from "./overviewDto";

describe("parseOverviewDto", () => {
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
      collected_at: "2026-07-14T12:00:00.000000000Z",
    };
  }

  it("accepts the canonical overview payload", () => {
    expect(parseOverviewDto(validPayload())).toEqual(validPayload());
  });

  it("accepts not_ready database state", () => {
    const payload = {
      ...validPayload(),
      readiness: {
        status: "not_ready",
        database: "error",
      },
    };
    expect(parseOverviewDto(payload)).toEqual(payload);
  });

  it("rejects unknown top-level fields", () => {
    expect(parseOverviewDto({ ...validPayload(), extra: "x" })).toBeNull();
  });

  it("rejects invalid readiness status", () => {
    expect(
      parseOverviewDto({
        ...validPayload(),
        readiness: { status: "degraded", database: "ok" },
      })
    ).toBeNull();
  });
});
