import { Suspense } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { PublicLayout } from "./layouts/PublicLayout";
import ApplianceApp from "../surfaces/appliance/ApplianceApp";
import { LoginRoute } from "../surfaces/auth/LoginRoute";
import { DownloadPageLegacy } from "../surfaces/public/download/DownloadPageLegacy";
import { LandingPageLegacy } from "../surfaces/public/landing/LandingPageLegacy";
import {
  DocsPreviewV2,
  DownloadPreviewV2,
  LandingPreviewV2,
  PreviewRouteFallback,
} from "../surfaces/public/preview/previewRoutes";
import { DocsPage, DownloadPage, LandingPage } from "../surfaces/public/publicRoutes";
import type { ApiClient } from "../lib/api";

type AppRouterProps = {
  apiClient?: ApiClient;
};

export function AppRouter({ apiClient }: AppRouterProps) {
  return (
    <Routes>
      <Route
        path="/"
        element={
          <Suspense fallback={<PreviewRouteFallback />}>
            <LandingPage />
          </Suspense>
        }
      />
      <Route
        path="/download"
        element={
          <Suspense fallback={<PreviewRouteFallback />}>
            <DownloadPage />
          </Suspense>
        }
      />
      <Route
        path="/docs"
        element={
          <Suspense fallback={<PreviewRouteFallback />}>
            <DocsPage />
          </Suspense>
        }
      />
      <Route element={<PublicLayout />}>
        <Route path="/design-preview/landing-legacy" element={<LandingPageLegacy />} />
        <Route path="/design-preview/download-legacy" element={<DownloadPageLegacy />} />
      </Route>
      <Route path="/login" element={<LoginRoute apiClient={apiClient} />} />
      <Route path="/app/*" element={<ApplianceApp apiClient={apiClient} />} />
      <Route
        path="/design-preview/landing"
        element={
          <Suspense fallback={<PreviewRouteFallback />}>
            <LandingPreviewV2 />
          </Suspense>
        }
      />
      <Route
        path="/design-preview/download"
        element={
          <Suspense fallback={<PreviewRouteFallback />}>
            <DownloadPreviewV2 />
          </Suspense>
        }
      />
      <Route
        path="/design-preview/docs"
        element={
          <Suspense fallback={<PreviewRouteFallback />}>
            <DocsPreviewV2 />
          </Suspense>
        }
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
