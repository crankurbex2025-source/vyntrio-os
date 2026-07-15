import { PublicControlSurfaceFrame, type PublicControlSurfaceFrameProps } from "./PublicControlSurfaceFrame";
import { PublicSectionIntro } from "./PublicSectionIntro";

export type PublicControlSurfaceShowcaseProps = {
  eyebrow?: string;
  sectionHeading: string;
  sectionDescription?: string;
  sectionHeadingId?: string;
  frameHeadingId?: string;
  framed?: boolean;
} & Omit<PublicControlSurfaceFrameProps, "headingId">;

export function PublicControlSurfaceShowcase({
  eyebrow,
  sectionHeading,
  sectionDescription,
  sectionHeadingId,
  frameHeadingId,
  framed = false,
  ...frameProps
}: PublicControlSurfaceShowcaseProps) {
  const frame = (
    <PublicControlSurfaceFrame {...frameProps} headingId={frameHeadingId} variant="showcase" />
  );

  return (
    <div className={framed ? "vyn-public-surface-showcase vyn-public-surface-showcase-framed" : "vyn-public-surface-showcase"}>
      <PublicSectionIntro
        eyebrow={eyebrow}
        heading={sectionHeading}
        description={sectionDescription}
        headingId={sectionHeadingId}
      />
      {framed ? <div className="vyn-public-surface-showcase-mount">{frame}</div> : frame}
    </div>
  );
}
