import { lazy } from "react";

export const LandingPage = lazy(() =>
  import("./landing/LandingPage").then((module) => ({ default: module.LandingPage }))
);

export const DownloadPage = lazy(() =>
  import("./download/DownloadPage").then((module) => ({ default: module.DownloadPage }))
);

export const DocsPage = lazy(() =>
  import("./docs/DocsPage").then((module) => ({ default: module.DocsPage }))
);
