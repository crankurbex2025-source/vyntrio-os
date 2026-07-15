import { I18nProvider, useI18n } from "../../../shared/i18n/I18nProvider";
import {
  PublicApplianceJourney,
  PublicDocsSection,
  PublicDownloadPanel,
  PublicHeroSection,
  PublicInlineCtaBand,
  PublicPreviewPageContext,
  PublicPreviewShell,
  PublicReleaseStrip,
  PublicSectionBand,
} from "../components";
import { PreviewPageMotion } from "./motion";
import "./motion/preview-motion.css";
import "./public-preview-product.css";
import { buildPreviewContextLinks } from "./previewContextConfig";
import { buildPreviewShellProps } from "./previewShellConfig";
import { usePreviewDocumentLang } from "./usePreviewDocumentLang";

function DownloadPreviewContent() {
  const { messages } = useI18n();
  usePreviewDocumentLang();

  return (
    <PublicPreviewShell {...buildPreviewShellProps(messages)} shellVariant="product">
      <div className="vyn-public-preview-page">
        <PreviewPageMotion>
          <PublicSectionBand tone="elevated" surface="route-hero">
            <PublicPreviewPageContext
              ariaLabel={messages.previewContext.ariaLabel}
              links={buildPreviewContextLinks(messages)}
              currentKey="download"
            />
            <PublicHeroSection
              eyebrow={messages.downloadPage.hero.eyebrow}
              title={messages.downloadPage.hero.title}
              description={messages.downloadPage.hero.description}
              titleId="preview-download-hero-title"
              variant="compact"
              accentLine
            />
          </PublicSectionBand>

          <PublicSectionBand surface="statement">
            <PublicReleaseStrip
              label={messages.release.label}
              title={messages.release.title}
              body={messages.release.body}
              titleId="preview-download-release-title"
            />
          </PublicSectionBand>

          <PublicSectionBand tone="inset" surface="journey">
            <PublicApplianceJourney
              eyebrow={messages.installJourney.eyebrow}
              heading={messages.installJourney.heading}
              intro={messages.installJourney.intro}
              steps={[...messages.installJourney.steps]}
              ariaLabel={messages.installJourney.ariaLabel}
              headingId="preview-download-journey-heading"
            />
          </PublicSectionBand>

          <PublicSectionBand tone="inset" surface="artifact">
            <PublicDownloadPanel
              heading={messages.downloadPage.panel.heading}
              intro={messages.downloadPage.panel.intro}
              rows={messages.downloadPage.panel.rows}
              note={messages.downloadPage.panel.note}
              headingId="preview-download-panel-heading"
            />
          </PublicSectionBand>

          <PublicSectionBand>
            <PublicDocsSection
              eyebrow={messages.downloadPage.requirements.eyebrow}
              heading={messages.downloadPage.requirements.heading}
              intro={messages.downloadPage.requirements.intro}
              items={messages.downloadPage.requirements.items}
              headingId="preview-download-requirements-heading"
            />
          </PublicSectionBand>

          <PublicSectionBand tone="elevated">
            <PublicInlineCtaBand
              heading={messages.downloadPage.inlineCta.heading}
              body={messages.downloadPage.inlineCta.body}
              primaryLabel={messages.downloadPage.inlineCta.primary}
              primaryTo="/design-preview/landing"
              secondaryLabel={messages.downloadPage.inlineCta.secondary}
              secondaryTo="/design-preview/docs"
              headingId="preview-download-inline-cta-heading"
            />
          </PublicSectionBand>
        </PreviewPageMotion>
      </div>
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
