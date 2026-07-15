import type { PublicMessages } from "../../../shared/i18n";
import type { PublicPreviewContextLink } from "../components/PublicPreviewPageContext";

export function buildProductionContextLinks(messages: PublicMessages): PublicPreviewContextLink[] {
  return [
    { key: "landing", label: messages.previewContext.landing, to: "/" },
    { key: "download", label: messages.previewContext.download, to: "/download" },
    { key: "docs", label: messages.previewContext.docs, to: "/design-preview/docs" },
  ];
}
