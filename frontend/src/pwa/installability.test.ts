import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";

const publicDir = join(import.meta.dirname, "../../public");

describe("PWA installability foundation (11R.7)", () => {
  it("ships a production web manifest with honest install boundaries", () => {
    const manifest = JSON.parse(
      readFileSync(join(publicDir, "site.webmanifest"), "utf8")
    ) as Record<string, unknown>;

    expect(manifest.name).toBe("Vyntrio");
    expect(manifest.short_name).toBe("Vyntrio");
    expect(manifest.start_url).toBe("/");
    expect(manifest.scope).toBe("/");
    expect(manifest.display).toBe("standalone");
    expect(manifest.background_color).toBe("#06080d");
    expect(manifest.theme_color).toBe("#06080d");
    expect(String(manifest.description)).toMatch(/not.*hardware/i);

    const icons = manifest.icons as Array<{ sizes: string; src: string }>;
    expect(icons.some((icon) => icon.sizes === "192x192")).toBe(true);
    expect(icons.some((icon) => icon.sizes === "512x512")).toBe(true);
    expect(icons.every((icon) => !icon.src.includes("design-preview"))).toBe(true);
  });

  it("includes required PNG icon assets for installability", () => {
    for (const size of [192, 512]) {
      const bytes = readFileSync(join(publicDir, "icons", `icon-${size}.png`));
      expect(bytes[0]).toBe(0x89);
      expect(bytes[1]).toBe(0x50);
      expect(bytes[2]).toBe(0x4e);
      expect(bytes[3]).toBe(0x47);
    }
  });
});
