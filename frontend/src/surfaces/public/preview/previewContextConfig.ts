import type { PublicMessages } from "../../../shared/i18n";
import type { PublicPreviewContextLink } from "../components/PublicPreviewPageContext";

export function buildPreviewContextLinks(messages: PublicMessages): PublicPreviewContextLink[] {
  return [
    { key: "landing", label: messages.previewContext.landing, to: "/design-preview/landing" },
    { key: "download", label: messages.previewContext.download, to: "/design-preview/download" },
    { key: "docs", label: messages.previewContext.docs, to: "/design-preview/docs" },
  ];
}
