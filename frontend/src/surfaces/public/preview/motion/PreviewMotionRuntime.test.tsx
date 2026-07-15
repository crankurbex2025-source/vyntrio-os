import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { PublicSectionBand } from "../../components/PublicSectionBand";
import { PreviewMotionRuntime } from "./PreviewMotionRuntime";

const mocks = vi.hoisted(() => ({
  gsapFrom: vi.fn(),
  registerPlugin: vi.fn(),
  scrollTriggerCreate: vi.fn(),
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
  ScrollTrigger: {
    create: (...args: unknown[]) => mocks.scrollTriggerCreate(...args),
  },
}));

vi.mock("@gsap/react", () => ({
  useGSAP: (callback: () => void) => {
    callback();
  },
}));

describe("PreviewMotionRuntime", () => {
  beforeEach(() => {
    mocks.gsapFrom.mockClear();
    mocks.registerPlugin.mockClear();
    mocks.scrollTriggerCreate.mockClear();
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

  it("renders motion scope and keeps children visible", () => {
    render(
      <PreviewMotionRuntime>
        <PublicSectionBand>
          <p>Runtime section</p>
        </PublicSectionBand>
      </PreviewMotionRuntime>
    );

    expect(screen.getByText("Runtime section")).toBeInTheDocument();
    expect(document.querySelector(".vyn-preview-motion-scope")).toHaveAttribute("data-motion", "on");
  });
});
