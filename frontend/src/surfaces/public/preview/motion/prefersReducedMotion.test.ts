import { describe, expect, it, vi, beforeEach, afterEach } from "vitest";
import { prefersReducedMotion, watchReducedMotion } from "./prefersReducedMotion";

describe("prefersReducedMotion", () => {
  const listeners = new Map<string, Set<() => void>>();

  beforeEach(() => {
    listeners.clear();
    vi.stubGlobal("matchMedia", (query: string) => ({
      matches: query === "(prefers-reduced-motion: reduce)",
      media: query,
      addEventListener: (_event: string, listener: () => void) => {
        const set = listeners.get(query) ?? new Set();
        set.add(listener);
        listeners.set(query, set);
      },
      removeEventListener: (_event: string, listener: () => void) => {
        listeners.get(query)?.delete(listener);
      },
    }));
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("returns true when prefers-reduced-motion is reduce", () => {
    expect(prefersReducedMotion()).toBe(true);
  });

  it("notifies watchReducedMotion subscribers on media change", () => {
    const handler = vi.fn();
    const unsubscribe = watchReducedMotion(handler);

    listeners.get("(prefers-reduced-motion: reduce)")?.forEach((listener) => listener());
    expect(handler).toHaveBeenCalledWith(true);

    unsubscribe();
  });
});
