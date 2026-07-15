import { useEffect } from "react";
import { useI18n } from "../../../shared/i18n/I18nProvider";

export function usePreviewDocumentLang(): void {
  const { locale } = useI18n();

  useEffect(() => {
    document.documentElement.lang = locale;
  }, [locale]);
}
