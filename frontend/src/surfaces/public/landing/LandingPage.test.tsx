import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { LandingPage } from "./LandingPage";

function installMediaNotBuiltResponse() {
  return {
    publication_status: "not_built",
    release: { version: "0.2.0-dev", channel: "development" },
    primary_artifact: {
      name: "vyntrio-install-media.img",
      format: "raw_gpt_hybrid_disk",
      firmware_boot_mode: "bios+uefi",
      download_available: false,
    },
    build_target: "make install-media",
    stage_target: "make release-install-media-stage",
    limitations: ["Dual-mode BIOS+UEFI required; BIOS-only incomplete"],
  };
}

describe("LandingPage", () => {
  beforeEach(() => {
    window.localStorage.removeItem("vyntrio.locale");
    window.localStorage.removeItem("vyntrio.public-theme");
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        if (url.includes("/api/v1/public/install-media")) {
          return new Response(JSON.stringify(installMediaNotBuiltResponse()), {
            status: 200,
            headers: { "Content-Type": "application/json" },
          });
        }
        return new Response("{}", { status: 404 });
      })
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("renders German production landing without preview banner", async () => {
    render(
      <MemoryRouter>
        <LandingPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "Vyntrio", level: 1 })).toBeInTheDocument();
    expect(screen.queryByText(/Design-Vorschau/i)).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Installationsmedien laden" })).toHaveAttribute(
      "href",
      "/download"
    );
    expect(screen.getByRole("link", { name: "Am Gerät anmelden" })).toHaveAttribute("href", "/login");
    await waitFor(() => {
      expect(screen.getByRole("heading", { name: "Installationsmedien und Checksummen" })).toBeInTheDocument();
    });
    expect(screen.getByRole("heading", { name: "Wo Container und VMs im Produkt stehen" })).toBeInTheDocument();
    expect(screen.queryByText("Checking session...")).not.toBeInTheDocument();
  });

  it("keeps a single Download CTA in primary chrome (no duplicate nav entries)", () => {
    render(
      <MemoryRouter>
        <LandingPage />
      </MemoryRouter>
    );

    const desktopNav = screen.getByRole("navigation", { name: "Hauptnavigation" });
    const desktopLabels = within(desktopNav)
      .getAllByRole("link")
      .map((link) => link.textContent?.trim());
    expect(desktopLabels).toEqual(["Produkt", "Storage", "Warum Vyntrio", "Docs"]);
    expect(desktopLabels.filter((label) => label === "Download")).toHaveLength(0);

    const downloadLinks = screen.getAllByRole("link", { name: "Download" });
    // Header CTA + footer only (hero uses a different label).
    expect(downloadLinks.length).toBeGreaterThanOrEqual(1);
    expect(downloadLinks.every((link) => link.getAttribute("href") === "/download")).toBe(true);
  });

  it("switches to English copy via locale switcher", async () => {
    render(
      <MemoryRouter>
        <LandingPage />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "EN" }));

    expect(screen.getByRole("heading", { name: "Vyntrio", level: 1 })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Download install media" })).toHaveAttribute(
      "href",
      "/download"
    );
    expect(
      screen.getByRole("heading", { name: "What works today vs what does not" })
    ).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "What you can do locally right now" })).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByRole("heading", { name: "Install media and checksums" })).toBeInTheDocument();
    });
  });

  it("renders local-today section in German by default", () => {
    render(
      <MemoryRouter>
        <LandingPage />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "Was du lokal schon nutzen kannst" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Menü öffnen" })).toBeInTheDocument();
  });
});
