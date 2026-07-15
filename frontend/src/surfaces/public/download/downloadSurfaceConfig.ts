import type { PublicPreviewContextLink } from "../components/PublicPreviewPageContext";

export type PublicDownloadSurfaceConfig = {
  mode: "production" | "preview";
  contextCurrentKey: "download";
  contextLinks: PublicPreviewContextLink[];
  inlineCtaPrimaryTo: string;
  inlineCtaSecondaryTo: string;
  idPrefix: string;
};
