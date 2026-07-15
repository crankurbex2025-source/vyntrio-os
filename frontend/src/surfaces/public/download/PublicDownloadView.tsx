import { useI18n } from "../../../shared/i18n/I18nProvider";
import {
  PublicApplianceJourney,
  PublicDocsSection,
  PublicDownloadPanel,
  PublicHeroSection,
  PublicInlineCtaBand,
  PublicPreviewPageContext,
  PublicProcedureOutline,
  PublicReleaseStrip,
  PublicSectionBand,
} from "../components";
import { PreviewPageMotion } from "../preview/motion";
import "../preview/motion/preview-motion.css";
import "../preview/public-preview-product.css";
import type { PublicDownloadSurfaceConfig } from "./downloadSurfaceConfig";

export type PublicDownloadViewProps = {
  surface: PublicDownloadSurfaceConfig;
};

export function PublicDownloadView({ surface }: PublicDownloadViewProps) {
  const { messages } = useI18n();
  const { idPrefix } = surface;
  const inlineCta =
    surface.mode === "production"
      ? messages.downloadPage.productionInlineCta
      : messages.downloadPage.inlineCta;
  const contextAriaLabel =
    surface.mode === "production"
      ? messages.productContext.ariaLabel
      : messages.previewContext.ariaLabel;

  return (
    <div className="vyn-public-preview-page">
      <PreviewPageMotion>
        <PublicSectionBand tone="elevated" surface="route-hero">
          <PublicPreviewPageContext
            ariaLabel={contextAriaLabel}
            links={surface.contextLinks}
            currentKey={surface.contextCurrentKey}
          />
          <PublicHeroSection
            eyebrow={messages.downloadPage.hero.eyebrow}
            title={messages.downloadPage.hero.title}
            description={messages.downloadPage.hero.description}
            titleId={`${idPrefix}-download-hero-title`}
            variant="compact"
            accentLine
          />
        </PublicSectionBand>

        <PublicSectionBand surface="statement">
          <PublicReleaseStrip
            label={messages.release.label}
            title={messages.release.title}
            body={messages.release.body}
            titleId={`${idPrefix}-download-release-title`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="inset" surface="journey">
          <PublicApplianceJourney
            eyebrow={messages.installJourney.eyebrow}
            heading={messages.installJourney.heading}
            intro={messages.installJourney.intro}
            steps={[...messages.installJourney.steps]}
            ariaLabel={messages.installJourney.ariaLabel}
            headingId={`${idPrefix}-download-journey-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="inset" surface="readiness">
          <PublicDocsSection
            eyebrow={messages.downloadPage.readiness.eyebrow}
            heading={messages.downloadPage.readiness.heading}
            intro={messages.downloadPage.readiness.intro}
            items={messages.downloadPage.readiness.items}
            headingId={`${idPrefix}-download-readiness-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="inset" surface="artifact">
          <PublicDownloadPanel
            heading={messages.downloadPage.panel.heading}
            intro={messages.downloadPage.panel.intro}
            rows={messages.downloadPage.panel.rows}
            note={messages.downloadPage.panel.note}
            headingId={`${idPrefix}-download-panel-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="inset">
          <PublicProcedureOutline
            eyebrow={messages.downloadPage.installOutline.eyebrow}
            heading={messages.downloadPage.installOutline.heading}
            intro={messages.downloadPage.installOutline.intro}
            steps={[...messages.downloadPage.installOutline.steps]}
            ariaLabel={messages.downloadPage.installOutline.ariaLabel}
            headingId={`${idPrefix}-download-install-outline-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand>
          <PublicDocsSection
            eyebrow={messages.downloadPage.mediaPrep.eyebrow}
            heading={messages.downloadPage.mediaPrep.heading}
            intro={messages.downloadPage.mediaPrep.intro}
            items={messages.downloadPage.mediaPrep.items}
            headingId={`${idPrefix}-download-media-prep-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand>
          <PublicDocsSection
            eyebrow={messages.downloadPage.requirements.eyebrow}
            heading={messages.downloadPage.requirements.heading}
            intro={messages.downloadPage.requirements.intro}
            items={messages.downloadPage.requirements.items}
            headingId={`${idPrefix}-download-requirements-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="elevated">
          <PublicInlineCtaBand
            heading={inlineCta.heading}
            body={inlineCta.body}
            primaryLabel={inlineCta.primary}
            primaryTo={surface.inlineCtaPrimaryTo}
            secondaryLabel={inlineCta.secondary}
            secondaryTo={surface.inlineCtaSecondaryTo}
            headingId={`${idPrefix}-download-inline-cta-heading`}
          />
        </PublicSectionBand>
      </PreviewPageMotion>
    </div>
  );
}
