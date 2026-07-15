import { useI18n } from "../i18n/I18nProvider";
import type { Locale } from "../i18n/types";

export function LocaleSwitcher() {
  const { locale, setLocale, messages } = useI18n();

  function handleSelect(next: Locale) {
    if (next !== locale) {
      setLocale(next);
    }
  }

  return (
    <div className="vyn-public-locale-switcher" role="group" aria-label="Language">
      <button
        type="button"
        className={locale === "de" ? "vyn-public-locale-switcher-active" : undefined}
        aria-pressed={locale === "de"}
        onClick={() => handleSelect("de")}
      >
        {messages.nav.localeDe}
      </button>
      <button
        type="button"
        className={locale === "en" ? "vyn-public-locale-switcher-active" : undefined}
        aria-pressed={locale === "en"}
        onClick={() => handleSelect("en")}
      >
        {messages.nav.localeEn}
      </button>
    </div>
  );
}
