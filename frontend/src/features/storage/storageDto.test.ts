import { describe, expect, it } from "vitest";
import {
  countEligibleDisks,
  formatExclusionReason,
  parseStorageDisksDto,
} from "./storageDto";

describe("parseStorageDisksDto", () => {
  it("parses a valid inventory response", () => {
    const payload = {
      collected_at: "2026-07-16T10:00:00.000000000Z",
      status: "ok",
      disks: [
        {
          id: "disk-abc123",
          status: "ok",
          size_bytes: 1000000000000,
          rotational: false,
          removable: false,
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
    };

    const parsed = parseStorageDisksDto(payload);
    expect(parsed).not.toBeNull();
    expect(parsed?.disks).toHaveLength(2);
    expect(countEligibleDisks(parsed!.disks)).toBe(1);
  });

  it("rejects unknown inventory status", () => {
    expect(
      parseStorageDisksDto({
        collected_at: "2026-07-16T10:00:00.000000000Z",
        status: "partial",
        disks: [],
      })
    ).toBeNull();
  });

  it("rejects invalid exclusion reason", () => {
    expect(
      parseStorageDisksDto({
        collected_at: "2026-07-16T10:00:00.000000000Z",
        status: "ok",
        disks: [
          {
            id: "disk-bad",
            status: "ok",
            eligibility: "excluded",
            reasons: ["not_a_real_reason"],
          },
        ],
      })
    ).toBeNull();
  });
});

describe("formatExclusionReason", () => {
  it("maps known reasons to readable labels", () => {
    expect(formatExclusionReason("root_disk")).toContain("root filesystem");
  });
});
