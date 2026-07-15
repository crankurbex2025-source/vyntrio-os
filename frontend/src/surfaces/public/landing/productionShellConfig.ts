import type { PublicMessages } from "../../../shared/i18n";
import type { PublicPreviewShellProps } from "../components/PublicPreviewShell";

type ProductionShellOptions = Pick<PublicPreviewShellProps, "anchorLinks" | "navAriaLabel">;

export function buildProductionShellProps(
  messages: PublicMessages,
  options: ProductionShellOptions = {}
): Omit<PublicPreviewShellProps, "children"> {
  return {
    brand: messages.brand,
    brandTo: "/",
    downloadLabel: messages.nav.download,
    downloadTo: "/download",
    signInLabel: messages.nav.signIn,
    signInTo: "/login",
    footerTagline: messages.footer.tagline,
    routeLinks: [
      { label: messages.nav.home, to: "/" },
      { label: messages.nav.download, to: "/download" },
      { label: messages.nav.docs, to: "/design-preview/docs" },
    ],
    footerLinks: [
      { label: messages.footer.download, to: "/download" },
      { label: messages.footer.docs, to: "/design-preview/docs" },
      { label: messages.footer.signIn, to: "/login" },
    ],
    navAriaLabel: options.navAriaLabel ?? "Public navigation",
    anchorLinks: options.anchorLinks,
    premiumSurface: true,
    shellVariant: "landing",
  };
}
