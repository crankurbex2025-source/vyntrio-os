import { I18nProvider, useI18n } from "../../../shared/i18n/I18nProvider";
import {
  PublicApplianceJourney,
  PublicDocsSection,
  PublicHeroSection,
  PublicInlineCtaBand,
  PublicPreviewPageContext,
  PublicPreviewShell,
  PublicProductStatusBlock,
  PublicReleaseStrip,
  PublicSectionBand,
} from "../components";
import { PreviewPageMotion } from "./motion";
import "./motion/preview-motion.css";
import "./public-preview-product.css";
import { buildPreviewContextLinks } from "./previewContextConfig";
import { buildPreviewShellProps } from "./previewShellConfig";
import { usePreviewDocumentLang } from "./usePreviewDocumentLang";

function DocsPreviewContent() {
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
              currentKey="docs"
            />
            <PublicHeroSection
              eyebrow={messages.docsPage.hero.eyebrow}
              title={messages.docsPage.hero.title}
              description={messages.docsPage.hero.description}
              titleId="preview-docs-hero-title"
              variant="compact"
              accentLine
            />
          </PublicSectionBand>

          <PublicSectionBand surface="statement">
            <PublicReleaseStrip
              label={messages.release.label}
              title={messages.release.title}
              body={messages.release.body}
              titleId="preview-docs-release-title"
            />
          </PublicSectionBand>

          <PublicSectionBand tone="inset" surface="journey">
            <PublicApplianceJourney
              eyebrow={messages.operateJourney.eyebrow}
              heading={messages.operateJourney.heading}
              intro={messages.operateJourney.intro}
              steps={[...messages.operateJourney.steps]}
              ariaLabel={messages.operateJourney.ariaLabel}
              headingId="preview-docs-journey-heading"
            />
          </PublicSectionBand>

          <PublicSectionBand tone="inset">
            {messages.docsPage.sections.map((section, index) => (
              <PublicDocsSection
                key={section.heading}
                eyebrow={section.eyebrow}
                heading={section.heading}
                intro={section.intro}
                items={section.items}
                headingId={`preview-docs-section-${index}-heading`}
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
              headingId="preview-docs-product-status-heading"
            />
          </PublicSectionBand>

          <PublicSectionBand tone="elevated">
            <PublicInlineCtaBand
              heading={messages.docsPage.inlineCta.heading}
              body={messages.docsPage.inlineCta.body}
              primaryLabel={messages.docsPage.inlineCta.primary}
              primaryTo="/design-preview/download"
              secondaryLabel={messages.docsPage.inlineCta.secondary}
              secondaryTo="/design-preview/landing"
              headingId="preview-docs-inline-cta-heading"
            />
          </PublicSectionBand>
        </PreviewPageMotion>
      </div>
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
