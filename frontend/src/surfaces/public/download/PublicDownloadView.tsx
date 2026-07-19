import { useI18n } from "../../../shared/i18n/I18nProvider";
import {
  PublicApplianceJourney,
  PublicDocsSection,
  PublicInlineCtaBand,
  PublicInstallMediaCreatorGuide,
  PublicInstallMediaSection,
  PublicInstallMediaWriterSection,
  PublicMediaCreatorPreview,
  PublicPlatformCreatorHero,
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
  const platform = messages.downloadPage.platformHero;

  return (
    <div className="vyn-public-preview-page">
      <PreviewPageMotion>
        <PublicSectionBand tone="elevated" surface="route-hero">
          <PublicPlatformCreatorHero
            eyebrow={platform.eyebrow}
            title={platform.title}
            description={platform.description}
            windowsCta={platform.windowsCta}
            macosCta={platform.macosCta}
            linuxCta={platform.linuxCta}
            macosBlockedNote={platform.macosBlockedNote}
            cardsHeading={platform.cardsHeading}
            cardsIntro={platform.cardsIntro}
            downloadUnavailable={messages.downloadPage.mediaWriter.downloadUnavailable}
            titleId={`${idPrefix}-download-hero-title`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="inset" surface="creator-preview">
          <PublicMediaCreatorPreview
            heading={messages.downloadPage.creatorPreview.heading}
            intro={messages.downloadPage.creatorPreview.intro}
            headingId={`${idPrefix}-download-creator-preview-heading`}
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

        <PublicSectionBand tone="inset" surface="artifact">
          <PublicInstallMediaSection
            heading={messages.downloadPage.installMedia.heading}
            intro={messages.downloadPage.installMedia.intro}
            buildHeading={messages.downloadPage.installMedia.buildHeading}
            buildIntro={messages.downloadPage.installMedia.buildIntro}
            downloadHeading={messages.downloadPage.installMedia.downloadHeading}
            limitationsHeading={messages.downloadPage.installMedia.limitationsHeading}
            statusLabels={messages.downloadPage.installMedia.statusLabels}
            rowLabels={messages.downloadPage.installMedia.rowLabels}
            headingId={`${idPrefix}-download-panel-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="inset" surface="media-writer">
          <PublicInstallMediaWriterSection
            heading={messages.downloadPage.mediaWriter.heading}
            intro={messages.downloadPage.mediaWriter.intro}
            honestNote={messages.downloadPage.mediaWriter.honestNote}
            flowHeading={messages.downloadPage.mediaWriter.flowHeading}
            flowSteps={[...messages.downloadPage.mediaWriter.flowSteps]}
            commandHeading={messages.downloadPage.mediaWriter.commandHeading}
            listCommand={messages.downloadPage.mediaWriter.listCommand}
            writeCommand={messages.downloadPage.mediaWriter.writeCommand}
            verifyHeading={messages.downloadPage.mediaWriter.verifyHeading}
            verifyWindows={messages.downloadPage.mediaWriter.verifyWindows}
            verifyMacOS={messages.downloadPage.mediaWriter.verifyMacOS}
            verifyLinux={messages.downloadPage.mediaWriter.verifyLinux}
            buildNote={messages.downloadPage.mediaWriter.buildNote}
            downloadsHeading={messages.downloadPage.mediaWriter.downloadsHeading}
            downloadsIntro={messages.downloadPage.mediaWriter.downloadsIntro}
            downloadUnavailable={messages.downloadPage.mediaWriter.downloadUnavailable}
            guiDownloadsHeading={messages.downloadPage.mediaWriter.guiDownloadsHeading}
            cliDownloadsHeading={messages.downloadPage.mediaWriter.cliDownloadsHeading}
            headingId={`${idPrefix}-download-media-writer-heading`}
          />
        </PublicSectionBand>

        <PublicSectionBand tone="inset" surface="media-creator">
          <PublicInstallMediaCreatorGuide
            heading={messages.downloadPage.mediaCreator.heading}
            intro={messages.downloadPage.mediaCreator.intro}
            earlyAccessNote={messages.downloadPage.mediaCreator.earlyAccessNote}
            artifactHeading={messages.downloadPage.mediaCreator.artifactHeading}
            warningsHeading={messages.downloadPage.mediaCreator.warningsHeading}
            warnings={[...messages.downloadPage.mediaCreator.warnings]}
            usbHeading={messages.downloadPage.mediaCreator.usbHeading}
            usbIntro={messages.downloadPage.mediaCreator.usbIntro}
            usbSteps={[...messages.downloadPage.mediaCreator.usbSteps]}
            usbHelperLabel={messages.downloadPage.mediaCreator.usbHelperLabel}
            usbHelperCommand={messages.downloadPage.mediaCreator.usbHelperCommand}
            vmHeading={messages.downloadPage.mediaCreator.vmHeading}
            vmIntro={messages.downloadPage.mediaCreator.vmIntro}
            vmSteps={[...messages.downloadPage.mediaCreator.vmSteps]}
            vmHelperLabel={messages.downloadPage.mediaCreator.vmHelperLabel}
            vmHelperCommand={messages.downloadPage.mediaCreator.vmHelperCommand}
            afterBootHeading={messages.downloadPage.mediaCreator.afterBootHeading}
            afterBootBody={messages.downloadPage.mediaCreator.afterBootBody}
            checksumPending={messages.downloadPage.mediaCreator.checksumPending}
            downloadPending={messages.downloadPage.mediaCreator.downloadPending}
            headingId={`${idPrefix}-download-media-creator-heading`}
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
