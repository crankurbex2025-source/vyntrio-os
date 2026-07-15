import { I18nProvider, useI18n } from "../../../shared/i18n/I18nProvider";
import { PublicPreviewShell } from "../components";
import { PublicLandingView } from "../landing/PublicLandingView";
import { buildPreviewContextLinks } from "./previewContextConfig";
import { buildPreviewShellProps } from "./previewShellConfig";
import { usePreviewDocumentLang } from "./usePreviewDocumentLang";

function LandingPreviewContent() {
  const { messages } = useI18n();
  usePreviewDocumentLang();

  return (
    <PublicPreviewShell
      {...buildPreviewShellProps(messages, {
        anchorLinks: [
          { label: messages.nav.home, to: "#product" },
          { label: messages.nav.capabilities, to: "#capabilities" },
        ],
      })}
      shellVariant="landing"
    >
      <PublicLandingView
        surface={{
          mode: "preview",
          contextCurrentKey: "landing",
          contextLinks: buildPreviewContextLinks(messages),
          ctaDownloadTo: "/design-preview/download",
          ctaSignInTo: "/login",
          finalCtaDownloadTo: "/design-preview/download",
          idPrefix: "preview",
        }}
      />
    </PublicPreviewShell>
  );
}

export function LandingPreviewV2() {
  return (
    <I18nProvider>
      <LandingPreviewContent />
    </I18nProvider>
  );
}
