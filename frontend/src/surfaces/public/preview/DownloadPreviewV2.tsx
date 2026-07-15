import { I18nProvider, useI18n } from "../../../shared/i18n/I18nProvider";
import {
  PublicDocsSection,
  PublicDownloadPanel,
  PublicHeroSection,
  PublicInlineCtaBand,
  PublicPreviewShell,
  PublicReleaseStrip,
  PublicSectionBand,
} from "../components";
import { buildPreviewShellProps } from "./previewShellConfig";
import { usePreviewDocumentLang } from "./usePreviewDocumentLang";

function DownloadPreviewContent() {
  const { messages } = useI18n();
  usePreviewDocumentLang();

  return (
    <PublicPreviewShell {...buildPreviewShellProps(messages)}>
      <PublicSectionBand tone="elevated">
        <PublicHeroSection
          eyebrow={messages.downloadPage.hero.eyebrow}
          title={messages.downloadPage.hero.title}
          description={messages.downloadPage.hero.description}
          titleId="preview-download-hero-title"
          variant="compact"
        />
      </PublicSectionBand>

      <PublicSectionBand>
        <PublicReleaseStrip
          label={messages.release.label}
          title={messages.release.title}
          body={messages.release.body}
          titleId="preview-download-release-title"
        />
      </PublicSectionBand>

      <PublicSectionBand tone="inset">
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
