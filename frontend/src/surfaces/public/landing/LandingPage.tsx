import { I18nProvider, useI18n } from "../../../shared/i18n/I18nProvider";
import { PublicPreviewShell } from "../components";
import { usePreviewDocumentLang } from "../preview/usePreviewDocumentLang";
import { PublicLandingView } from "./PublicLandingView";
import { buildProductionShellProps } from "./productionShellConfig";

function LandingPageContent() {
  const { messages } = useI18n();
  usePreviewDocumentLang();

  return (
    <PublicPreviewShell {...buildProductionShellProps(messages, { shellVariant: "landing" })}>
      <PublicLandingView
        surface={{
          mode: "production",
          contextCurrentKey: "landing",
          contextLinks: [],
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
