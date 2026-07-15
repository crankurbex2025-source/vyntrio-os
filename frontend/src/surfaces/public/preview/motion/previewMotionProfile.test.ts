import { describe, expect, it, vi, afterEach } from "vitest";
import { getPreviewMotionProfile } from "./previewMotionConfig";

describe("getPreviewMotionProfile", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns reduced offsets on narrow viewports", () => {
    vi.stubGlobal("matchMedia", (query: string) => ({
      matches: query === "(max-width: 40rem)",
      media: query,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }));

    const profile = getPreviewMotionProfile();
    expect(profile.companionOffsetX).toBe(0);
    expect(profile.heroOffsetY).toBeLessThan(18);
  });

  it("returns desktop offsets on wide viewports", () => {
    vi.stubGlobal("matchMedia", () => ({
      matches: false,
      media: "",
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }));

    const profile = getPreviewMotionProfile();
    expect(profile.companionOffsetX).toBe(28);
    expect(profile.heroOffsetY).toBe(18);
  });
});
