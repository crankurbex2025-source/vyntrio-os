import type { PublicPreviewContextLink } from "../components/PublicPreviewPageContext";

export type PublicDocsSurfaceConfig = {
  mode: "production" | "preview";
  contextCurrentKey: "docs";
  contextLinks: PublicPreviewContextLink[];
  inlineCtaPrimaryTo: string;
  inlineCtaSecondaryTo: string;
  idPrefix: string;
};
