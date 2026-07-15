import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { PublicSectionBand } from "../../components/PublicSectionBand";
import { PreviewPageMotion } from "./PreviewPageMotion";

const mocks = vi.hoisted(() => ({
  gsapFrom: vi.fn(),
  registerPlugin: vi.fn(),
}));

vi.mock("gsap", () => ({
  default: {
    registerPlugin: mocks.registerPlugin,
    from: (...args: unknown[]) => mocks.gsapFrom(...args),
    utils: {
      toArray: (selector: string, root?: ParentNode | null) => {
        const scope = root ?? document;
        return Array.from(scope.querySelectorAll(selector));
      },
    },
  },
}));

vi.mock("gsap/ScrollTrigger", () => ({
  ScrollTrigger: {},
}));

vi.mock("@gsap/react", () => ({
  useGSAP: (callback: () => void) => {
    callback();
  },
}));

describe("PreviewPageMotion", () => {
  beforeEach(() => {
    mocks.gsapFrom.mockClear();
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

  it("renders children and marks motion on when reduced motion is not preferred", () => {
    render(
      <PreviewPageMotion>
        <PublicSectionBand>
          <p>Section content</p>
        </PublicSectionBand>
      </PreviewPageMotion>
    );

    expect(screen.getByText("Section content")).toBeInTheDocument();
    expect(document.querySelector(".vyn-preview-motion-scope")).toHaveAttribute("data-motion", "on");
  });

  it("skips GSAP setup when prefers-reduced-motion is enabled", () => {
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
    expect(mocks.gsapFrom).not.toHaveBeenCalled();
  });
});
