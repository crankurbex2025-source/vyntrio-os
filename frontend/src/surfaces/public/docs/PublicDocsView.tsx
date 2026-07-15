import { useI18n } from "../../../shared/i18n/I18nProvider";
import {
  PublicApplianceJourney,
  PublicDocsGuideBand,
  PublicDocsSection,
  PublicHeroSection,
  PublicInlineCtaBand,
  PublicPreviewPageContext,
  PublicProductStatusBlock,
  PublicReleaseStrip,
  PublicSectionBand,
} from "../components";
import { PreviewPageMotion } from "../preview/motion";
import "../preview/motion/preview-motion.css";
import "../preview/public-preview-product.css";
import type { PublicDocsSurfaceConfig } from "./docsSurfaceConfig";

export type PublicDocsViewProps = {
  surface: PublicDocsSurfaceConfig;
};

export function PublicDocsView({ surface }: PublicDocsViewProps) {
  const { messages } = useI18n();
  const { idPrefix } = surface;
  const inlineCta =
    surface.mode === "production" ? messages.docsPage.productionInlineCta : messages.docsPage.inlineCta;
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
            eyebrow={messages.docsPage.hero.eyebrow}
            title={messages.docsPage.hero.title}
            description={messages.docsPage.hero.description}
            titleId={`${idPrefix}-docs-hero-title`}
            variant="compact"
            accentLine
          />
        </PublicSectionBand>

        <PublicSectionBand surface="statement">
          <PublicReleaseStrip
            label={messages.release.label}
            title={messages.release.title}
            body={messages.release.body}
            titleId={`${idPrefix}-docs-release-title`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="inset" surface="journey">
          <PublicApplianceJourney
            eyebrow={messages.operateJourney.eyebrow}
            heading={messages.operateJourney.heading}
            intro={messages.operateJourney.intro}
            steps={[...messages.operateJourney.steps]}
            ariaLabel={messages.operateJourney.ariaLabel}
            headingId={`${idPrefix}-docs-journey-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="inset">
          <PublicDocsGuideBand
            eyebrow={messages.docsPage.guide.eyebrow}
            heading={messages.docsPage.guide.heading}
            body={messages.docsPage.guide.body}
            headingId={`${idPrefix}-docs-guide-heading`}
          />
          {messages.docsPage.sections.map((section, index) => (
            <PublicDocsSection
              key={section.heading}
              eyebrow={section.eyebrow}
              heading={section.heading}
              intro={section.intro}
              items={section.items}
              headingId={`${idPrefix}-docs-section-${index}-heading`}
            />
          ))}
        </PublicSectionBand>

        <PublicSectionBand>
          <PublicProductStatusBlock
            variant="terminal"
            eyebrow={messages.productStatus.eyebrow}
            heading={messages.productStatus.heading}
            body={messages.productStatus.body}
            points={[...messages.productStatus.points]}
            headingId={`${idPrefix}-docs-product-status-heading`}
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
            headingId={`${idPrefix}-docs-inline-cta-heading`}
          />
        </PublicSectionBand>
      </PreviewPageMotion>
    </div>
  );
}
