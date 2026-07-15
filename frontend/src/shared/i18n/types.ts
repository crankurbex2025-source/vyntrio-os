export const SUPPORTED_LOCALES = ["de", "en"] as const;

export type Locale = (typeof SUPPORTED_LOCALES)[number];

export const DEFAULT_LOCALE: Locale = "de";

export type PublicMessages = {
  previewBanner: string;
  brand: string;
  nav: {
    home: string;
    capabilities: string;
    download: string;
    signIn: string;
    localeDe: string;
    localeEn: string;
  };
  hero: {
    eyebrow: string;
    title: string;
    description: string;
    ctaDownload: string;
    ctaDownloadHint: string;
    ctaSignIn: string;
    ctaSignInHint: string;
  };
  release: {
    label: string;
    title: string;
    body: string;
  };
  pillars: {
    eyebrow: string;
    heading: string;
    intro: string;
    storage: { tag: string; title: string; body: string };
    services: { tag: string; title: string; body: string };
    control: { tag: string; title: string; body: string };
  };
  surfaceShowcase: {
    eyebrow: string;
    heading: string;
    intro: string;
  };
  surfacePreview: {
    heading: string;
    subheading: string;
    panelLabel: string;
    panelNote: string;
    rows: Array<{ label: string; value: string }>;
  };
  productStatus: {
    eyebrow: string;
    heading: string;
    body: string;
    points: [string, string, string];
  };
  finalCta: {
    heading: string;
    body: string;
    ctaDownload: string;
    ctaSignIn: string;
  };
  footer: {
    tagline: string;
    download: string;
    signIn: string;
    previewNote: string;
  };
};

export type MessageCatalog = {
  public: PublicMessages;
};
