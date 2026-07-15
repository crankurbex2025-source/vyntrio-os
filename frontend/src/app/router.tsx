import { Suspense } from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import { PublicLayout } from "./layouts/PublicLayout";
import ApplianceApp from "../surfaces/appliance/ApplianceApp";
import { LoginRoute } from "../surfaces/auth/LoginRoute";
import { DownloadPlaceholder } from "../surfaces/public/download/DownloadPlaceholder";
import { LandingPage } from "../surfaces/public/landing/LandingPage";
import {
  DocsPreviewV2,
  DownloadPreviewV2,
  LandingPreviewV2,
  PreviewRouteFallback,
} from "../surfaces/public/preview/previewRoutes";
import type { ApiClient } from "../lib/api";

type AppRouterProps = {
  apiClient?: ApiClient;
};

export function AppRouter({ apiClient }: AppRouterProps) {
  return (
    <Routes>
      <Route element={<PublicLayout />}>
        <Route path="/" element={<LandingPage />} />
        <Route path="/download" element={<DownloadPlaceholder />} />
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
