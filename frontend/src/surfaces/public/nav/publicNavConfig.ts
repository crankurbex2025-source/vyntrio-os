import type { PublicMessages } from "../../../shared/i18n";

export type PublicNavItem = {
  id: string;
  label: string;
  to: string;
  cta?: boolean;
  secondary?: boolean;
};

export type PublicNavModel = {
  brand: string;
  brandTo: string;
  items: PublicNavItem[];
  cta: PublicNavItem;
  secondary: PublicNavItem;
  footerItems: PublicNavItem[];
  footerTagline: string;
  footerNote?: string;
};

type NavPathOptions = {
  homeTo: string;
  downloadTo: string;
  docsTo: string;
  signInTo?: string;
  capabilitiesTo: string;
  whyTo: string;
  homeLabel?: "home" | "previewLanding";
  footerNote?: string;
};

/**
 * Single source of truth for public header + footer navigation.
 * Download appears once as the primary CTA — never duplicated in text nav.
 */
export function buildPublicNavModel(
  messages: PublicMessages,
  paths: NavPathOptions
): PublicNavModel {
  const homeLabel =
    paths.homeLabel === "previewLanding"
      ? messages.nav.previewLanding
      : messages.nav.home;

  const items: PublicNavItem[] = [
    { id: "home", label: homeLabel, to: paths.homeTo },
    {
      id: "capabilities",
      label: messages.nav.capabilities,
      to: paths.capabilitiesTo,
    },
    { id: "why", label: messages.nav.useCases, to: paths.whyTo },
    { id: "docs", label: messages.nav.docs, to: paths.docsTo },
  ];

  const cta: PublicNavItem = {
    id: "download",
    label: messages.nav.download,
    to: paths.downloadTo,
    cta: true,
  };

  const secondary: PublicNavItem = {
    id: "sign-in",
    label: messages.nav.signIn,
    to: paths.signInTo ?? "/login",
    secondary: true,
  };

  return {
    brand: messages.brand,
    brandTo: paths.homeTo,
    items,
    cta,
    secondary,
    footerItems: [
      { id: "footer-download", label: messages.footer.download, to: paths.downloadTo },
      { id: "footer-docs", label: messages.footer.docs, to: paths.docsTo },
      { id: "footer-why", label: messages.nav.useCases, to: paths.whyTo },
      { id: "footer-sign-in", label: messages.footer.signIn, to: secondary.to },
    ],
    footerTagline: messages.footer.tagline,
    footerNote: paths.footerNote,
  };
}

export function buildProductionNav(messages: PublicMessages): PublicNavModel {
  return buildPublicNavModel(messages, {
    homeTo: "/",
    downloadTo: "/download",
    docsTo: "/docs",
    signInTo: "/login",
    capabilitiesTo: "/#storage",
    whyTo: "/#use-cases",
  });
}

export function buildPreviewNav(messages: PublicMessages): PublicNavModel {
  return buildPublicNavModel(messages, {
    homeTo: "/design-preview/landing",
    downloadTo: "/design-preview/download",
    docsTo: "/design-preview/docs",
    signInTo: "/login",
    capabilitiesTo: "/design-preview/landing#storage",
    whyTo: "/design-preview/landing#use-cases",
    homeLabel: "previewLanding",
    footerNote: messages.footer.previewNote,
  });
}
