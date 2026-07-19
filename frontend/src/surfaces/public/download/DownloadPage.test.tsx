import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { DownloadPage } from "./DownloadPage";

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

describe("DownloadPage", () => {
  beforeEach(() => {
    window.localStorage.removeItem("vyntrio.locale");
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

  it("renders German production download without preview banner or API dependencies", async () => {
    render(
      <MemoryRouter>
        <DownloadPage />
      </MemoryRouter>
    );

    expect(
      screen.getByRole("heading", { name: "Media Creator für deine Plattform laden" })
    ).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getAllByText("vyntrio-install-media.img").length).toBeGreaterThan(0);
    });
    expect(screen.getAllByText("make install-media").length).toBeGreaterThan(0);
    expect(
      screen.getByRole("heading", { name: "Native Media-Creator-Pakete" })
    ).toBeInTheDocument();
    expect(screen.queryByText(/Design-Vorschau/i)).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Zur Startseite" })).toHaveAttribute("href", "/");
    expect(screen.getByRole("link", { name: "Zur Dokumentation" })).toHaveAttribute("href", "/docs");
  });

  it("switches to English copy via locale switcher", async () => {
    render(
      <MemoryRouter>
        <DownloadPage />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "EN" }));

    expect(screen.getByRole("heading", { name: "Download Media Creator for your platform" })).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getAllByText("vyntrio-install-media.img").length).toBeGreaterThan(0);
    });
    expect(screen.getByRole("heading", { name: "Native Media Creator packages" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Platform availability" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Native desktop packages" })).toBeInTheDocument();
    expect(screen.getAllByText("Not staged on this host — build from source").length).toBeGreaterThan(0);
    expect(screen.getByRole("heading", { name: "Create bootable install media" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Go to homepage" })).toHaveAttribute("href", "/");
  });
});
