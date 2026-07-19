import { useI18n } from "../../../shared/i18n/I18nProvider";
import {
  PublicApplianceChassisArt,
  PublicApplianceJourney,
  PublicAppsVmsSection,
  PublicControlSurfaceShowcase,
  PublicFinalCtaBand,
  PublicHeroLayout,
  PublicHeroSection,
  PublicLiveReleaseBand,
  PublicPillarSection,
  PublicProcedureOutline,
  PublicProductStatusBlock,
  PublicSectionBand,
  PublicTrustBand,
  PublicUseCaseSection,
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

  return (
    <div className="vyn-public-landing-page vyn-public-preview-page">
      <PreviewPageMotion variant="landing">
        {/* 1–2. Hero: brand-first composition */}
        <PublicSectionBand tone="elevated" surface="hero" id="product">
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
            ariaLabelledBy={`${idPrefix}-hero-title`}
          />
        </PublicSectionBand>

        {/* 3. Download / live release */}
        <PublicSectionBand tone="inset" surface="artifact" id="download">
          <PublicLiveReleaseBand
            eyebrow={messages.liveRelease.eyebrow}
            heading={messages.liveRelease.heading}
            intro={messages.liveRelease.intro}
            downloadCta={messages.liveRelease.downloadCta}
            downloadTo={surface.ctaDownloadTo}
            headingId={`${idPrefix}-live-release-heading`}
          />
        </PublicSectionBand>

        {/* 4. USB Creator / install path */}
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

        {/* 5. Core platform pillars */}
        <PublicSectionBand tone="panel" surface="enterprise" id="storage">
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

        <PublicSectionBand tone="trust" surface="trust" id="trust">
          <PublicTrustBand
            heading={messages.trustBand.heading}
            headingEmphasis={messages.trustBand.headingEmphasis}
            intro={messages.trustBand.intro}
            marks={[...messages.trustBand.marks]}
            headingId={`${idPrefix}-trust-band-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="inset" surface="capabilities" id="apps-vms">
          <PublicAppsVmsSection
            eyebrow={messages.appsVms.eyebrow}
            heading={messages.appsVms.heading}
            intro={messages.appsVms.intro}
            appsTitle={messages.appsVms.appsTitle}
            appsBody={messages.appsVms.appsBody}
            appsStatus={messages.appsVms.appsStatus}
            vmsTitle={messages.appsVms.vmsTitle}
            vmsBody={messages.appsVms.vmsBody}
            vmsStatus={messages.appsVms.vmsStatus}
            headingId={`${idPrefix}-apps-vms-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="inset" surface="capabilities" id="use-cases">
          <PublicUseCaseSection
            eyebrow={messages.useCases.eyebrow}
            heading={messages.useCases.heading}
            intro={messages.useCases.intro}
            ariaLabel={messages.useCases.ariaLabel}
            cases={[...messages.useCases.cases]}
            headingId={`${idPrefix}-use-cases-heading`}
          />
        </PublicSectionBand>

        {/* 6. Product truth / status */}
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

        <PublicSectionBand tone="inset" id="local-today">
          <PublicProcedureOutline
            eyebrow={messages.localToday.eyebrow}
            heading={messages.localToday.heading}
            intro={messages.localToday.intro}
            ariaLabel={messages.localToday.ariaLabel}
            steps={[...messages.localToday.steps]}
            headingId={`${idPrefix}-local-today-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="inset" id="first-boot-setup">
          <PublicProcedureOutline
            eyebrow={messages.firstBootSetup.eyebrow}
            heading={messages.firstBootSetup.heading}
            intro={messages.firstBootSetup.intro}
            ariaLabel={messages.firstBootSetup.ariaLabel}
            steps={[...messages.firstBootSetup.steps]}
            headingId={`${idPrefix}-first-boot-setup-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="panel" surface="enterprise" id="control-surface">
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

        {/* 7–8. Docs / ecosystem CTA + footer handled by shell */}
        <PublicSectionBand tone="elevated" surface="finale" id="docs-support">
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
