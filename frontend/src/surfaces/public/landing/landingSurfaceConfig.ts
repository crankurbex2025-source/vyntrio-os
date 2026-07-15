import type { PublicPreviewContextLink } from "../components/PublicPreviewPageContext";

export type PublicLandingSurfaceConfig = {
  mode: "production" | "preview";
  contextCurrentKey: string;
  contextLinks: PublicPreviewContextLink[];
  ctaDownloadTo: string;
  ctaSignInTo: string;
  finalCtaDownloadTo: string;
  idPrefix: string;
};
