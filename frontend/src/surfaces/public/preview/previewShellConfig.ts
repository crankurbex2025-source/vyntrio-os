import type { PublicMessages } from "../../../shared/i18n";
import type { PublicPreviewShellProps } from "../components/PublicPreviewShell";

type PreviewShellOptions = Pick<PublicPreviewShellProps, "anchorLinks" | "navAriaLabel">;

export function buildPreviewShellProps(
  messages: PublicMessages,
  options: PreviewShellOptions = {}
): Omit<PublicPreviewShellProps, "children"> {
  return {
    banner: messages.previewBanner,
    brand: messages.brand,
    brandTo: "/design-preview/landing",
    downloadLabel: messages.nav.download,
    downloadTo: "/design-preview/download",
    signInLabel: messages.nav.signIn,
    footerTagline: messages.footer.tagline,
    footerNote: messages.footer.previewNote,
    routeLinks: [
      { label: messages.nav.previewLanding, to: "/design-preview/landing" },
      { label: messages.nav.download, to: "/design-preview/download" },
      { label: messages.nav.docs, to: "/design-preview/docs" },
    ],
    footerLinks: [
      { label: messages.footer.download, to: "/design-preview/download" },
      { label: messages.footer.docs, to: "/design-preview/docs" },
      { label: messages.footer.signIn, to: "/login" },
    ],
    navAriaLabel: options.navAriaLabel ?? "Preview navigation",
    anchorLinks: options.anchorLinks,
    premiumSurface: true,
  };
}
