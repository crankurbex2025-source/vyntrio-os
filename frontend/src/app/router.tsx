import { Navigate, Route, Routes } from "react-router-dom";
import { PublicLayout } from "./layouts/PublicLayout";
import ApplianceApp from "../surfaces/appliance/ApplianceApp";
import { LoginRoute } from "../surfaces/auth/LoginRoute";
import { DownloadPlaceholder } from "../surfaces/public/download/DownloadPlaceholder";
import { LandingPage } from "../surfaces/public/landing/LandingPage";
import { LandingPreviewV2 } from "../surfaces/public/preview/LandingPreviewV2";
import { DownloadPreviewV2 } from "../surfaces/public/preview/DownloadPreviewV2";
import { DocsPreviewV2 } from "../surfaces/public/preview/DocsPreviewV2";
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
      <Route path="/design-preview/landing" element={<LandingPreviewV2 />} />
      <Route path="/design-preview/download" element={<DownloadPreviewV2 />} />
      <Route path="/design-preview/docs" element={<DocsPreviewV2 />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
