import { Link } from "react-router-dom";
import { PageHeader } from "../../../shared/ui/PageHeader";

/** Slice 11.1 download placeholder — retained at /design-preview/download-legacy for rollback review. */
export function DownloadPageLegacy() {
  return (
    <main className="public-main download-placeholder">
      <PageHeader
        eyebrow="Download"
        title="Install media coming soon"
        description="Vyntrio OS install images are not publicly available in this build yet."
      />
      <p className="download-placeholder-note" role="status">
        There is no download file linked from this page. When release media is ready, this route
        will point to verified install artifacts and documentation — not placeholder binaries.
      </p>
      <div className="landing-hero-actions">
        <Link className="public-button public-button-secondary" to="/">
          Back to home
        </Link>
        <Link className="public-button public-button-secondary" to="/login">
          Sign in to an existing appliance
        </Link>
      </div>
    </main>
  );
}
