import { I18nProvider, useI18n } from "../../../shared/i18n/I18nProvider";
import { PublicPreviewShell } from "../components";
import { PublicDownloadView } from "../download/PublicDownloadView";
import { buildPreviewContextLinks } from "./previewContextConfig";
import { buildPreviewShellProps } from "./previewShellConfig";
import { usePreviewDocumentLang } from "./usePreviewDocumentLang";

function DownloadPreviewContent() {
  const { messages } = useI18n();
  usePreviewDocumentLang();

  return (
    <PublicPreviewShell {...buildPreviewShellProps(messages)} shellVariant="product">
      <PublicDownloadView
        surface={{
          mode: "preview",
          contextCurrentKey: "download",
          contextLinks: buildPreviewContextLinks(messages),
          inlineCtaPrimaryTo: "/design-preview/landing",
          inlineCtaSecondaryTo: "/design-preview/docs",
          idPrefix: "preview",
        }}
      />
    </PublicPreviewShell>
  );
}

export function DownloadPreviewV2() {
  return (
    <I18nProvider>
      <DownloadPreviewContent />
    </I18nProvider>
  );
}
