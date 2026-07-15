import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import { getPublicMessages } from "../../../shared/i18n";
import { I18nProvider } from "../../../shared/i18n/I18nProvider";
import {
  PublicControlSurfaceFrame,
  PublicControlSurfaceShowcase,
  PublicFinalCtaBand,
  PublicHeroSection,
  PublicPillarSection,
  PublicPreviewShell,
  PublicProductStatusBlock,
  PublicReleaseStrip,
  PublicSectionBand,
} from "./index";

const de = getPublicMessages("de");

describe("public surface components", () => {
  it("PublicHeroSection renders CTA hints without API dependencies", () => {
    render(
      <MemoryRouter>
        <PublicHeroSection
          eyebrow={de.hero.eyebrow}
          title={de.hero.title}
          description={de.hero.description}
          ctaDownloadLabel={de.hero.ctaDownload}
          ctaDownloadHint={de.hero.ctaDownloadHint}
          ctaSignInLabel={de.hero.ctaSignIn}
          ctaSignInHint={de.hero.ctaSignInHint}
        />
      </MemoryRouter>
    );

    expect(screen.getByRole("heading", { name: de.hero.title })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: de.hero.ctaDownload })).toHaveAttribute("href", "/download");
    expect(screen.getByText(de.hero.ctaDownloadHint)).toBeInTheDocument();
  });

  it("PublicControlSurfaceFrame shows honest static rows", () => {
    render(
      <PublicControlSurfaceFrame
        heading={de.surfacePreview.heading}
        subheading={de.surfacePreview.subheading}
        panelLabel={de.surfacePreview.panelLabel}
        panelNote={de.surfacePreview.panelNote}
        rows={de.surfacePreview.rows}
      />
    );

    expect(screen.getByText("Nach Login")).toBeInTheDocument();
    expect(screen.getByText("—")).toBeInTheDocument();
  });

  it("PublicPillarSection renders tagged unequal pillar set", () => {
    render(
      <PublicPillarSection
        eyebrow={de.pillars.eyebrow}
        heading={de.pillars.heading}
        intro={de.pillars.intro}
        pillars={[
          {
            tag: de.pillars.storage.tag,
            title: de.pillars.storage.title,
            body: de.pillars.storage.body,
            featured: true,
          },
          {
            tag: de.pillars.services.tag,
            title: de.pillars.services.title,
            body: de.pillars.services.body,
          },
        ]}
      />
    );

    expect(screen.getByRole("heading", { name: de.pillars.heading })).toBeInTheDocument();
    expect(screen.getByText(de.pillars.storage.tag)).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: de.pillars.storage.title })).toBeInTheDocument();
  });

  it("PublicControlSurfaceShowcase composes intro and frame", () => {
    render(
      <PublicControlSurfaceShowcase
        eyebrow={de.surfaceShowcase.eyebrow}
        sectionHeading={de.surfaceShowcase.heading}
        sectionDescription={de.surfaceShowcase.intro}
        heading={de.surfacePreview.heading}
        subheading={de.surfacePreview.subheading}
        panelLabel={de.surfacePreview.panelLabel}
        panelNote={de.surfacePreview.panelNote}
        rows={de.surfacePreview.rows}
      />
    );

    expect(screen.getByRole("heading", { name: de.surfaceShowcase.heading })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: de.surfacePreview.heading })).toBeInTheDocument();
  });

  it("PublicPreviewShell includes locale switcher and footer links", () => {
    render(
      <MemoryRouter>
        <I18nProvider initialLocale="de">
          <PublicPreviewShell
            banner={de.previewBanner}
            brand={de.brand}
            downloadLabel={de.nav.download}
            signInLabel={de.nav.signIn}
            footerTagline={de.footer.tagline}
            footerNote={de.footer.previewNote}
            footerLinks={[
              { label: de.footer.download, to: "/download" },
              { label: de.footer.signIn, to: "/login" },
            ]}
          >
            <PublicSectionBand>
              <PublicReleaseStrip
                label={de.release.label}
                title={de.release.title}
                body={de.release.body}
              />
              <PublicProductStatusBlock
                eyebrow={de.productStatus.eyebrow}
                heading={de.productStatus.heading}
                body={de.productStatus.body}
                points={[...de.productStatus.points]}
              />
              <PublicFinalCtaBand
                heading={de.finalCta.heading}
                body={de.finalCta.body}
                ctaDownloadLabel={de.finalCta.ctaDownload}
                ctaSignInLabel={de.finalCta.ctaSignIn}
              />
            </PublicSectionBand>
          </PublicPreviewShell>
        </I18nProvider>
      </MemoryRouter>
    );

    expect(screen.getByRole("group", { name: "Language" })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: de.release.title })).toBeInTheDocument();
    expect(screen.getByRole("contentinfo").querySelector(".vyn-public-footer-nav")).toBeTruthy();
    expect(screen.getAllByRole("link", { name: de.footer.download })[0]).toHaveAttribute("href", "/download");
  });
});
