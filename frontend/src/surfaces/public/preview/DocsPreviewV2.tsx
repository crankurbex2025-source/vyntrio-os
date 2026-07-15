import { I18nProvider, useI18n } from "../../../shared/i18n/I18nProvider";
import {
  PublicDocsSection,
  PublicHeroSection,
  PublicInlineCtaBand,
  PublicPreviewShell,
  PublicProductStatusBlock,
  PublicReleaseStrip,
  PublicSectionBand,
} from "../components";
import { PreviewPageMotion } from "./motion";
import "./motion/preview-motion.css";
import { buildPreviewShellProps } from "./previewShellConfig";
import { usePreviewDocumentLang } from "./usePreviewDocumentLang";

function DocsPreviewContent() {
  const { messages } = useI18n();
  usePreviewDocumentLang();

  return (
    <PublicPreviewShell {...buildPreviewShellProps(messages)}>
      <PreviewPageMotion>
      <PublicSectionBand tone="elevated">
        <PublicHeroSection
          eyebrow={messages.docsPage.hero.eyebrow}
          title={messages.docsPage.hero.title}
          description={messages.docsPage.hero.description}
          titleId="preview-docs-hero-title"
          variant="compact"
        />
      </PublicSectionBand>

      <PublicSectionBand>
        <PublicReleaseStrip
          label={messages.release.label}
          title={messages.release.title}
          body={messages.release.body}
          titleId="preview-docs-release-title"
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
