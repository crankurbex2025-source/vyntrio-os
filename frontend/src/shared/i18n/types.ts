export type PublicResourceItem = {
  title: string;
  description: string;
  statusLabel?: string;
};

export type PublicDocsSectionMessages = {
  eyebrow?: string;
  heading: string;
  intro?: string;
  items: PublicResourceItem[];
};

export type PublicDownloadRowMessages = {
  label: string;
  value: string;
};

export const SUPPORTED_LOCALES = ["de", "en"] as const;

export type Locale = (typeof SUPPORTED_LOCALES)[number];

export const DEFAULT_LOCALE: Locale = "de";

export type PublicMessages = {
  previewBanner: string;
  brand: string;
  nav: {
    previewLanding: string;
    home: string;
    capabilities: string;
    docs: string;
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
  heroStory: {
    ariaLabel: string;
    steps: Array<{ label: string; detail: string }>;
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
    bezel: {
      powerLabel: string;
      linkLabel: string;
    };
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
  downloadPage: {
    hero: {
      eyebrow: string;
      title: string;
      description: string;
    };
    panel: {
      heading: string;
      intro: string;
      rows: PublicDownloadRowMessages[];
      note: string;
    };
    requirements: {
      eyebrow: string;
      heading: string;
      intro: string;
      items: PublicResourceItem[];
    };
    inlineCta: {
      heading: string;
      body: string;
      primary: string;
      secondary: string;
    };
  };
  docsPage: {
    hero: {
      eyebrow: string;
      title: string;
      description: string;
    };
    sections: [PublicDocsSectionMessages, PublicDocsSectionMessages];
    inlineCta: {
      heading: string;
      body: string;
      primary: string;
      secondary: string;
    };
  };
  footer: {
    tagline: string;
    download: string;
    docs: string;
    signIn: string;
    previewNote: string;
  };
};

export type MessageCatalog = {
  public: PublicMessages;
};
