import { I18nProvider, useI18n } from "../../../shared/i18n/I18nProvider";
import { PublicPreviewShell } from "../components";
import { PublicDocsView } from "../docs/PublicDocsView";
import { buildPreviewShellProps } from "./previewShellConfig";
import { usePreviewDocumentLang } from "./usePreviewDocumentLang";

function DocsPreviewContent() {
  const { messages } = useI18n();
  usePreviewDocumentLang();

  return (
    <PublicPreviewShell {...buildPreviewShellProps(messages)} shellVariant="product">
      <PublicDocsView
        surface={{
          mode: "preview",
          contextCurrentKey: "docs",
          contextLinks: [],
          inlineCtaPrimaryTo: "/design-preview/download",
          inlineCtaSecondaryTo: "/design-preview/landing",
          idPrefix: "preview",
        }}
      />
    </PublicPreviewShell>
  );
}

export function DocsPreviewV2() {
  return (
    <I18nProvider>
      <DocsPreviewContent />
    </I18nProvider>
  );
}
