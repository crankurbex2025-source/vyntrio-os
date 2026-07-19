import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { LandingPreviewV2 } from "./LandingPreviewV2";

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

describe("LandingPreviewV2", () => {
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

  it("renders German preview by default", async () => {
    render(
      <MemoryRouter>
        <LandingPreviewV2 />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: "Vyntrio", level: 1 })).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent(/Design-Vorschau/i);
    expect(screen.getAllByText(/Nach Login/i).length).toBeGreaterThan(0);
    expect(screen.getByRole("heading", { name: "Die Verwaltungsoberfläche im Überblick" })).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByRole("heading", { name: "Installationsmedien und Checksummen" })).toBeInTheDocument();
    });
    expect(screen.queryByText("Checking session...")).not.toBeInTheDocument();
  });

  it("switches to English copy via locale switcher", async () => {
    render(
      <MemoryRouter>
        <LandingPreviewV2 />
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole("button", { name: "EN" }));

    expect(screen.getByRole("heading", { name: "Vyntrio", level: 1 })).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent(/Design preview/i);
    expect(
      screen.getByRole("heading", { name: "What works today vs what does not" })
    ).toBeInTheDocument();
  });
});
