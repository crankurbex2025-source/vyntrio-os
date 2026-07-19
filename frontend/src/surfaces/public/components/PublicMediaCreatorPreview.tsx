import type { ReactNode } from "react";

/** Original framed preview of the native Media Creator flow (not a third-party screenshot). */
export function PublicMediaCreatorPreview({
  heading,
  intro,
  headingId = "public-media-creator-preview-heading",
}: {
  heading: string;
  intro: string;
  headingId?: string;
}) {
  return (
    <section className="vyn-public-creator-preview" aria-labelledby={headingId}>
      <h2 id={headingId}>{heading}</h2>
      <p className="vyn-public-download-panel-intro">{intro}</p>
      <div className="vyn-public-creator-preview-frame" role="img" aria-label="Vyntrio Media Creator flow preview">
        <div className="vyn-public-creator-preview-titlebar">
          <span>Vyntrio Media Creator</span>
          <span className="vyn-public-creator-preview-dots" aria-hidden="true">
            ▢ ▢ ▢
          </span>
        </div>
        <div className="vyn-public-creator-preview-body">
          <PreviewPane title="1. Welcome">Prepare bootable install media</PreviewPane>
          <PreviewPane title="2. Release">0.2.0-dev · bios+uefi · dual-mode</PreviewPane>
          <PreviewPane title="3. Storage">Select removable USB</PreviewPane>
          <PreviewPane title="4. Write">Confirm · progress · verify</PreviewPane>
        </div>
      </div>
    </section>
  );
}

function PreviewPane({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="vyn-public-creator-preview-pane">
      <strong>{title}</strong>
      <span>{children}</span>
    </div>
  );
}
