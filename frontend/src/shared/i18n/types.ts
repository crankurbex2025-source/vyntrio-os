export type PublicResourceItem = {
  title: string;
  description: string;
  statusLabel?: string;
  to?: string;
};

export type PublicProcedureStepMessages = {
  step: string;
  title: string;
  body: string;
};

export type PublicDocsGuideMessages = {
  eyebrow: string;
  heading: string;
  body: string;
};

export type PublicDocsSectionBlockMessages = {
  eyebrow?: string;
  heading: string;
  intro?: string;
  items: PublicResourceItem[];
};

export type PublicDocsSectionMessages = PublicDocsSectionBlockMessages;

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
  chrome: {
    navAriaLabel: string;
    menuOpen: string;
    menuClose: string;
    mobileNav: string;
    themeToggle: string;
  };
  nav: {
    previewLanding: string;
    home: string;
    capabilities: string;
    useCases: string;
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
  trustBand: {
    heading: string;
    headingEmphasis: string;
    intro: string;
    marks: Array<{ label: string; detail?: string }>;
  };
  previewContext: {
    ariaLabel: string;
    landing: string;
    download: string;
    docs: string;
  };
  productContext: {
    ariaLabel: string;
  };
  liveRelease: {
    eyebrow: string;
    heading: string;
    intro: string;
    downloadCta: string;
  };
  appsVms: {
    eyebrow: string;
    heading: string;
    intro: string;
    appsTitle: string;
    appsBody: string;
    appsStatus: string;
    vmsTitle: string;
    vmsBody: string;
    vmsStatus: string;
  };
  installJourney: {
    eyebrow: string;
    heading: string;
    intro: string;
    ariaLabel: string;
    steps: Array<{ phase: string; title: string; body: string; status: string }>;
  };
  operateJourney: {
    eyebrow: string;
    heading: string;
    intro: string;
    ariaLabel: string;
    steps: Array<{ phase: string; title: string; body: string; status: string }>;
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
  useCases: {
    eyebrow: string;
    heading: string;
    intro: string;
    ariaLabel: string;
    cases: Array<{ tag: string; title: string; body: string }>;
  };
  creatorAudience: {
    eyebrow: string;
    heading: string;
    intro: string;
    ariaLabel: string;
    cases: Array<{ tag: string; title: string; body: string }>;
  };
  localToday: {
    eyebrow: string;
    heading: string;
    intro: string;
    ariaLabel: string;
    steps: Array<{ step: string; title: string; body: string }>;
  };
  firstBootSetup: {
    eyebrow: string;
    heading: string;
    intro: string;
    ariaLabel: string;
    steps: Array<{ step: string; title: string; body: string }>;
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
    platformHero: {
      eyebrow: string;
      title: string;
      description: string;
      windowsCta: string;
      macosCta: string;
      linuxCta: string;
      macosBlockedNote: string;
      cardsHeading: string;
      cardsIntro: string;
    };
    creatorPreview: {
      heading: string;
      intro: string;
    };
    panel: {
      heading: string;
      intro: string;
      rows: PublicDownloadRowMessages[];
      note: string;
    };
    installMedia: {
      heading: string;
      intro: string;
      buildHeading: string;
      buildIntro: string;
      downloadHeading: string;
      limitationsHeading: string;
      statusLabels: {
        notBuilt: string;
        localStaging: string;
        unavailable: string;
        loading: string;
        error: string;
      };
      rowLabels: {
        installImage: string;
        imageFormat: string;
        checksum: string;
        verification: string;
        releaseChannel: string;
        buildTarget: string;
        stageTarget: string;
        download: string;
        generatedAt: string;
        supportStatus: string;
      };
    };
    mediaWriter: {
      heading: string;
      intro: string;
      honestNote: string;
      flowHeading: string;
      flowSteps: string[];
      commandHeading: string;
      listCommand: string;
      writeCommand: string;
      verifyHeading: string;
      verifyWindows: string;
      verifyMacOS: string;
      verifyLinux: string;
      buildNote: string;
      downloadsHeading: string;
      downloadsIntro: string;
      downloadUnavailable: string;
      guiDownloadsHeading: string;
      cliDownloadsHeading: string;
    };
    readiness: {
      eyebrow: string;
      heading: string;
      intro: string;
      items: PublicResourceItem[];
    };
    mediaCreator: {
      heading: string;
      intro: string;
      earlyAccessNote: string;
      artifactHeading: string;
      warningsHeading: string;
      warnings: string[];
      usbHeading: string;
      usbIntro: string;
      usbSteps: PublicProcedureStepMessages[];
      usbHelperLabel: string;
      usbHelperCommand: string;
      vmHeading: string;
      vmIntro: string;
      vmSteps: PublicProcedureStepMessages[];
      vmHelperLabel: string;
      vmHelperCommand: string;
      afterBootHeading: string;
      afterBootBody: string;
      checksumPending: string;
      downloadPending: string;
    };
    installOutline: {
      eyebrow: string;
      heading: string;
      intro: string;
      ariaLabel: string;
      steps: PublicProcedureStepMessages[];
    };
    mediaPrep: {
      eyebrow: string;
      heading: string;
      intro: string;
      items: PublicResourceItem[];
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
    productionInlineCta: {
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
    guide: PublicDocsGuideMessages;
    sections: PublicDocsSectionMessages[];
    inlineCta: {
      heading: string;
      body: string;
      primary: string;
      secondary: string;
    };
    productionInlineCta: {
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
