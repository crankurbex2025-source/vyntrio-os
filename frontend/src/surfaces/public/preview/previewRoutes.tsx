import { lazy, type ReactNode } from "react";

export const LandingPreviewV2 = lazy(() =>
  import("./LandingPreviewV2").then((module) => ({ default: module.LandingPreviewV2 }))
);

export const DownloadPreviewV2 = lazy(() =>
  import("./DownloadPreviewV2").then((module) => ({ default: module.DownloadPreviewV2 }))
);

export const DocsPreviewV2 = lazy(() =>
  import("./DocsPreviewV2").then((module) => ({ default: module.DocsPreviewV2 }))
);

export function PreviewRouteFallback(): ReactNode {
  return (
    <main
      aria-busy="true"
      aria-label="Loading design preview"
      style={{
        minHeight: "100vh",
        background: "#06080d",
      }}
    />
  );
}
