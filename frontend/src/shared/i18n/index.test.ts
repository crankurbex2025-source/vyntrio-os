import { describe, expect, it } from "vitest";
import { DEFAULT_LOCALE, getPublicMessages, isLocale, resolveInitialLocale } from "./index";

const STORAGE_KEY = "vyntrio.locale";

describe("i18n", () => {
  it("defaults to German catalog", () => {
    expect(DEFAULT_LOCALE).toBe("de");
    expect(getPublicMessages("de").hero.title).toContain("Hardware");
  });

  it("provides separate English copy, not a mirror of German strings", () => {
    const de = getPublicMessages("de");
    const en = getPublicMessages("en");
    expect(en.hero.title).not.toBe(de.hero.title);
    expect(en.hero.title.toLowerCase()).toContain("hardware");
  });

  it("validates supported locales", () => {
    expect(isLocale("de")).toBe(true);
    expect(isLocale("en")).toBe(true);
    expect(isLocale("fr")).toBe(false);
  });

  it("resolveInitialLocale uses persisted locale when set", () => {
    window.localStorage.setItem(STORAGE_KEY, "en");
    expect(resolveInitialLocale()).toBe("en");
    window.localStorage.removeItem(STORAGE_KEY);
    expect(resolveInitialLocale()).toBe("de");
  });
});
