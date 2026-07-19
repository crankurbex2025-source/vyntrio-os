import { I18nProvider, useI18n } from "../../../shared/i18n/I18nProvider";
import { PublicPreviewShell } from "../components";
import { buildProductionShellProps } from "../landing/productionShellConfig";
import { usePreviewDocumentLang } from "../preview/usePreviewDocumentLang";
import { PublicDownloadView } from "./PublicDownloadView";

function DownloadPageContent() {
  const { messages } = useI18n();
  usePreviewDocumentLang();

  return (
    <PublicPreviewShell
      {...buildProductionShellProps(messages, { shellVariant: "product" })}
    >
      <PublicDownloadView
        surface={{
          mode: "production",
          contextCurrentKey: "download",
          contextLinks: [],
          inlineCtaPrimaryTo: "/",
          inlineCtaSecondaryTo: "/docs",
          idPrefix: "public",
        }}
      />
    </PublicPreviewShell>
  );
}

export function DownloadPage() {
  return (
    <I18nProvider>
      <DownloadPageContent />
    </I18nProvider>
  );
}
