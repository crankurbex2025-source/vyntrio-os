import { I18nProvider, useI18n } from "../../../shared/i18n/I18nProvider";
import { PublicPreviewShell } from "../components";
import { buildProductionContextLinks } from "../landing/productionContextConfig";
import { buildProductionShellProps } from "../landing/productionShellConfig";
import { usePreviewDocumentLang } from "../preview/usePreviewDocumentLang";
import { PublicDocsView } from "./PublicDocsView";

function DocsPageContent() {
  const { messages } = useI18n();
  usePreviewDocumentLang();

  return (
    <PublicPreviewShell
      {...buildProductionShellProps(messages, { shellVariant: "product" })}
    >
      <PublicDocsView
        surface={{
          mode: "production",
          contextCurrentKey: "docs",
          contextLinks: buildProductionContextLinks(messages),
          inlineCtaPrimaryTo: "/download",
          inlineCtaSecondaryTo: "/",
          idPrefix: "public",
        }}
      />
    </PublicPreviewShell>
  );
}

export function DocsPage() {
  return (
    <I18nProvider>
      <DocsPageContent />
    </I18nProvider>
  );
}
