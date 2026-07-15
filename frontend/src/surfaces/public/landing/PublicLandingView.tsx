import { useI18n } from "../../../shared/i18n/I18nProvider";
import {
  PublicApplianceChassisArt,
  PublicApplianceJourney,
  PublicApplianceSignalPath,
  PublicControlSurfaceFrame,
  PublicControlSurfaceShowcase,
  PublicFinalCtaBand,
  PublicHeroLayout,
  PublicHeroSection,
  PublicPillarSection,
  PublicPreviewPageContext,
  PublicProductStatusBlock,
  PublicReleaseStrip,
  PublicSectionBand,
} from "../components";
import { PreviewPageMotion } from "../preview/motion";
import "../preview/motion/preview-motion.css";
import "../preview/public-landing-visual.css";
import "../preview/public-preview-product.css";
import type { PublicLandingSurfaceConfig } from "./landingSurfaceConfig";

export type PublicLandingViewProps = {
  surface: PublicLandingSurfaceConfig;
};

export function PublicLandingView({ surface }: PublicLandingViewProps) {
  const { messages } = useI18n();
  const { idPrefix } = surface;
  const contextAriaLabel =
    surface.mode === "production"
      ? messages.productContext.ariaLabel
      : messages.previewContext.ariaLabel;

  return (
    <div className="vyn-public-landing-page vyn-public-preview-page">
      <PreviewPageMotion variant="landing">
        <PublicSectionBand tone="elevated" surface="hero" id="product">
          <PublicPreviewPageContext
            ariaLabel={contextAriaLabel}
            links={surface.contextLinks}
            currentKey={surface.contextCurrentKey}
          />
          <PublicApplianceSignalPath
            steps={[...messages.heroStory.steps]}
            ariaLabel={messages.heroStory.ariaLabel}
          />
          <PublicHeroLayout
            premium
            art={<PublicApplianceChassisArt />}
            hero={
              <PublicHeroSection
                eyebrow={messages.hero.eyebrow}
                title={messages.hero.title}
                description={messages.hero.description}
                ctaDownloadLabel={messages.hero.ctaDownload}
                ctaDownloadHint={messages.hero.ctaDownloadHint}
                ctaDownloadTo={surface.ctaDownloadTo}
                ctaSignInLabel={messages.hero.ctaSignIn}
                ctaSignInHint={messages.hero.ctaSignInHint}
                ctaSignInTo={surface.ctaSignInTo}
                titleId={`${idPrefix}-hero-title`}
                variant="lead"
                accentLine
              />
            }
            companion={
              <PublicControlSurfaceFrame
                heading={messages.surfacePreview.heading}
                subheading={messages.surfacePreview.subheading}
                panelLabel={messages.surfacePreview.panelLabel}
                panelNote={messages.surfacePreview.panelNote}
                rows={messages.surfacePreview.rows}
                headingId={`${idPrefix}-surface-companion-heading`}
                chassis
                bezel={messages.surfacePreview.bezel}
              />
            }
            ariaLabelledBy={`${idPrefix}-hero-title`}
          />
        </PublicSectionBand>

        <PublicSectionBand surface="statement">
          <PublicReleaseStrip
            label={messages.release.label}
            title={messages.release.title}
            body={messages.release.body}
            titleId={`${idPrefix}-release-title`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="inset" surface="journey" id="install-journey">
          <PublicApplianceJourney
            eyebrow={messages.installJourney.eyebrow}
            heading={messages.installJourney.heading}
            intro={messages.installJourney.intro}
            steps={[...messages.installJourney.steps]}
            ariaLabel={messages.installJourney.ariaLabel}
            headingId={`${idPrefix}-install-journey-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="inset" surface="capabilities" id="capabilities">
          <PublicPillarSection
            eyebrow={messages.pillars.eyebrow}
            heading={messages.pillars.heading}
            intro={messages.pillars.intro}
            showGlyphs
            pillars={[
              {
                tag: messages.pillars.storage.tag,
                title: messages.pillars.storage.title,
                body: messages.pillars.storage.body,
                featured: true,
                glyph: "storage",
              },
              {
                tag: messages.pillars.services.tag,
                title: messages.pillars.services.title,
                body: messages.pillars.services.body,
                glyph: "services",
              },
              {
                tag: messages.pillars.control.tag,
                title: messages.pillars.control.title,
                body: messages.pillars.control.body,
                glyph: "control",
              },
            ]}
            headingId={`${idPrefix}-pillars-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="elevated" id="control-surface">
          <PublicControlSurfaceShowcase
            framed
            chassis
            bezel={messages.surfacePreview.bezel}
            eyebrow={messages.surfaceShowcase.eyebrow}
            sectionHeading={messages.surfaceShowcase.heading}
            sectionDescription={messages.surfaceShowcase.intro}
            sectionHeadingId={`${idPrefix}-surface-showcase-heading`}
            frameHeadingId={`${idPrefix}-surface-showcase-frame-heading`}
            heading={messages.surfacePreview.heading}
            subheading={messages.surfacePreview.subheading}
            panelLabel={messages.surfacePreview.panelLabel}
            panelNote={messages.surfacePreview.panelNote}
            rows={messages.surfacePreview.rows}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="inset" id="product-status">
          <PublicProductStatusBlock
            variant="terminal"
            eyebrow={messages.productStatus.eyebrow}
            heading={messages.productStatus.heading}
            body={messages.productStatus.body}
            points={[...messages.productStatus.points]}
            headingId={`${idPrefix}-product-status-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="elevated" surface="finale">
          <PublicFinalCtaBand
            heading={messages.finalCta.heading}
            body={messages.finalCta.body}
            ctaDownloadLabel={messages.finalCta.ctaDownload}
            ctaSignInLabel={messages.finalCta.ctaSignIn}
            ctaDownloadTo={surface.finalCtaDownloadTo}
            ctaSignInTo={surface.ctaSignInTo}
            headingId={`${idPrefix}-final-cta-heading`}
          />
        </PublicSectionBand>
      </PreviewPageMotion>
    </div>
  );
}
