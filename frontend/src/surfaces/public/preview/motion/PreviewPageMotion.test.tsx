import "@testing-library/jest-dom/vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { PublicSectionBand } from "../../components/PublicSectionBand";
import { PreviewPageMotion } from "./PreviewPageMotion";

vi.mock("./PreviewMotionRuntime", () => ({
  PreviewMotionRuntime: ({
    children,
    variant,
  }: {
    children: React.ReactNode;
    variant?: string;
  }) => (
    <div className="vyn-preview-motion-scope" data-motion="on" data-motion-variant={variant}>
      {children}
    </div>
  ),
}));

describe("PreviewPageMotion", () => {
  beforeEach(() => {
    vi.stubGlobal("matchMedia", (query: string) => ({
      matches: false,
      media: query,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }));
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("renders children and marks motion on when reduced motion is not preferred", async () => {
    render(
      <PreviewPageMotion>
        <PublicSectionBand>
          <p>Section content</p>
        </PublicSectionBand>
      </PreviewPageMotion>
    );

    expect(screen.getByText("Section content")).toBeInTheDocument();
    await waitFor(() => {
      expect(document.querySelector(".vyn-preview-motion-scope")).toHaveAttribute("data-motion", "on");
    });
  });

  it("marks landing motion variant when requested", async () => {
    render(
      <PreviewPageMotion variant="landing">
        <PublicSectionBand>
          <p>Landing section</p>
        </PublicSectionBand>
      </PreviewPageMotion>
    );

    await waitFor(() => {
      expect(document.querySelector(".vyn-preview-motion-scope")).toHaveAttribute(
        "data-motion-variant",
        "landing"
      );
    });
  });

  it("skips lazy GSAP runtime when prefers-reduced-motion is enabled", () => {
    vi.stubGlobal("matchMedia", (query: string) => ({
      matches: query === "(prefers-reduced-motion: reduce)",
      media: query,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }));

    render(
      <PreviewPageMotion>
        <PublicSectionBand>
          <p>Reduced motion section</p>
        </PublicSectionBand>
      </PreviewPageMotion>
    );

    expect(document.querySelector(".vyn-preview-motion-scope")).toHaveAttribute("data-motion", "off");
  });
});
