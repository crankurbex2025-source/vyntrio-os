import { describe, expect, it } from "vitest";
import { applianceSectionFromPath, buildApplianceNavItems } from "./applianceNavConfig";

describe("applianceNavConfig", () => {
  it("exposes primary Unraid-class sections without Apps/VMs", () => {
    const items = buildApplianceNavItems();
    expect(items.map((item) => item.id)).toEqual([
      "dashboard",
      "storage",
      "shares",
      "users",
      "settings",
      "tools",
    ]);
    expect(items.find((item) => item.id === "users")?.planned).toBe(true);
    expect(items.find((item) => item.id === "tools")?.planned).toBe(true);
    expect(items.some((item) => item.id === "apps" || item.id === "vms")).toBe(false);
  });

  it("maps paths to section ids", () => {
    expect(applianceSectionFromPath("/app")).toBe("dashboard");
    expect(applianceSectionFromPath("/app/")).toBe("dashboard");
    expect(applianceSectionFromPath("/app/storage")).toBe("storage");
    expect(applianceSectionFromPath("/app/shares")).toBe("shares");
    expect(applianceSectionFromPath("/app/users")).toBe("users");
    expect(applianceSectionFromPath("/app/settings")).toBe("settings");
    expect(applianceSectionFromPath("/app/tools")).toBe("tools");
  });
});
