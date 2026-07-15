import { I18nProvider, useI18n } from "../../../shared/i18n/I18nProvider";
import { PublicPreviewShell } from "../components";
import { usePreviewDocumentLang } from "../preview/usePreviewDocumentLang";
import { PublicLandingView } from "./PublicLandingView";
import { buildProductionContextLinks } from "./productionContextConfig";
import { buildProductionShellProps } from "./productionShellConfig";

function LandingPageContent() {
  const { messages } = useI18n();
  usePreviewDocumentLang();

  return (
    <PublicPreviewShell
      {...buildProductionShellProps(messages, {
        anchorLinks: [
          { label: messages.nav.home, to: "#product" },
          { label: messages.nav.capabilities, to: "#capabilities" },
        ],
      })}
    >
      <PublicLandingView
        surface={{
          mode: "production",
          contextCurrentKey: "landing",
          contextLinks: buildProductionContextLinks(messages),
          ctaDownloadTo: "/download",
          ctaSignInTo: "/login",
          finalCtaDownloadTo: "/download",
          idPrefix: "public",
        }}
      />
    </PublicPreviewShell>
  );
}

export function LandingPage() {
  return (
    <I18nProvider>
      <LandingPageContent />
    </I18nProvider>
  );
}
