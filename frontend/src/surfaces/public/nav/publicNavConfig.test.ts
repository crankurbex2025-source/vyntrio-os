import { describe, expect, it } from "vitest";
import { getPublicMessages } from "../../../shared/i18n";
import { buildProductionNav, buildPreviewNav } from "./publicNavConfig";

describe("publicNavConfig", () => {
  it("never duplicates Download in production text nav", () => {
    const nav = buildProductionNav(getPublicMessages("en"));
    expect(nav.items.map((item) => item.id)).toEqual(["home", "capabilities", "why", "docs"]);
    expect(nav.cta.to).toBe("/download");
    expect(nav.items.some((item) => item.to === "/download")).toBe(false);
  });

  it("keeps preview routes coherent without duplicate Download", () => {
    const nav = buildPreviewNav(getPublicMessages("en"));
    expect(nav.items.map((item) => item.to)).not.toContain("/design-preview/download");
    expect(nav.cta.to).toBe("/design-preview/download");
  });
});
