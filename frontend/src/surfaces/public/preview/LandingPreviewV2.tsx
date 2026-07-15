import { I18nProvider, useI18n } from "../../../shared/i18n/I18nProvider";
import {
  PublicControlSurfaceFrame,
  PublicControlSurfaceShowcase,
  PublicFinalCtaBand,
  PublicHeroLayout,
  PublicHeroSection,
  PublicPillarSection,
  PublicPreviewShell,
  PublicProductStatusBlock,
  PublicReleaseStrip,
  PublicSectionBand,
} from "../components";
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
    >
      <PublicSectionBand tone="elevated" id="product">
        <PublicHeroLayout
          hero={
            <PublicHeroSection
              eyebrow={messages.hero.eyebrow}
              title={messages.hero.title}
              description={messages.hero.description}
              ctaDownloadLabel={messages.hero.ctaDownload}
              ctaDownloadHint={messages.hero.ctaDownloadHint}
              ctaSignInLabel={messages.hero.ctaSignIn}
              ctaSignInHint={messages.hero.ctaSignInHint}
              titleId="preview-hero-title"
              variant="lead"
            />
          }
          companion={
            <PublicControlSurfaceFrame
              heading={messages.surfacePreview.heading}
              subheading={messages.surfacePreview.subheading}
              panelLabel={messages.surfacePreview.panelLabel}
              panelNote={messages.surfacePreview.panelNote}
              rows={messages.surfacePreview.rows}
              headingId="preview-surface-companion-heading"
            />
          }
          ariaLabelledBy="preview-hero-title"
        />
      </PublicSectionBand>

      <PublicSectionBand>
        <PublicReleaseStrip
          label={messages.release.label}
          title={messages.release.title}
          body={messages.release.body}
          titleId="preview-release-title"
        />
      </PublicSectionBand>

      <PublicSectionBand tone="inset" id="capabilities">
        <PublicPillarSection
          eyebrow={messages.pillars.eyebrow}
          heading={messages.pillars.heading}
          intro={messages.pillars.intro}
          pillars={[
            {
              tag: messages.pillars.storage.tag,
              title: messages.pillars.storage.title,
              body: messages.pillars.storage.body,
              featured: true,
            },
            {
              tag: messages.pillars.services.tag,
              title: messages.pillars.services.title,
              body: messages.pillars.services.body,
            },
            {
              tag: messages.pillars.control.tag,
              title: messages.pillars.control.title,
              body: messages.pillars.control.body,
            },
          ]}
          headingId="preview-pillars-heading"
        />
      </PublicSectionBand>

      <PublicSectionBand tone="elevated" id="control-surface">
        <PublicControlSurfaceShowcase
          eyebrow={messages.surfaceShowcase.eyebrow}
          sectionHeading={messages.surfaceShowcase.heading}
          sectionDescription={messages.surfaceShowcase.intro}
          sectionHeadingId="preview-surface-showcase-heading"
          frameHeadingId="preview-surface-showcase-frame-heading"
          heading={messages.surfacePreview.heading}
          subheading={messages.surfacePreview.subheading}
          panelLabel={messages.surfacePreview.panelLabel}
          panelNote={messages.surfacePreview.panelNote}
          rows={messages.surfacePreview.rows}
        />
      </PublicSectionBand>

      <PublicSectionBand tone="inset" id="product-status">
        <PublicProductStatusBlock
          eyebrow={messages.productStatus.eyebrow}
          heading={messages.productStatus.heading}
          body={messages.productStatus.body}
          points={[...messages.productStatus.points]}
          headingId="preview-product-status-heading"
        />
      </PublicSectionBand>

      <PublicSectionBand tone="elevated">
        <PublicFinalCtaBand
          heading={messages.finalCta.heading}
          body={messages.finalCta.body}
          ctaDownloadLabel={messages.finalCta.ctaDownload}
          ctaSignInLabel={messages.finalCta.ctaSignIn}
          ctaDownloadTo="/design-preview/download"
          headingId="preview-final-cta-heading"
        />
      </PublicSectionBand>
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
