import dePublic from "./locales/de/public.json";
import enPublic from "./locales/en/public.json";
import {
  DEFAULT_LOCALE,
  SUPPORTED_LOCALES,
  type Locale,
  type MessageCatalog,
  type PublicMessages,
} from "./types";

const STORAGE_KEY = "vyntrio.locale";

const catalogs: Record<Locale, MessageCatalog> = {
  de: { public: dePublic as PublicMessages },
  en: { public: enPublic as PublicMessages },
};

export function isLocale(value: string): value is Locale {
  return (SUPPORTED_LOCALES as readonly string[]).includes(value);
}

export function resolveInitialLocale(): Locale {
  if (typeof window === "undefined") {
    return DEFAULT_LOCALE;
  }

  const stored = window.localStorage.getItem(STORAGE_KEY);
  if (stored && isLocale(stored)) {
    return stored;
  }

  // German-first product default — no Accept-Language auto-switch until URL locale routes (11R.3).
  return DEFAULT_LOCALE;
}

export function persistLocale(locale: Locale): void {
  window.localStorage.setItem(STORAGE_KEY, locale);
}

export function getPublicMessages(locale: Locale): PublicMessages {
  return catalogs[locale].public;
}

export { DEFAULT_LOCALE, SUPPORTED_LOCALES };
export type { Locale, PublicMessages };
