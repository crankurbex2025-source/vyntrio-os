import { PublicControlSurfaceFrame, type PublicControlSurfaceFrameProps } from "./PublicControlSurfaceFrame";
import { PublicSectionIntro } from "./PublicSectionIntro";

export type PublicControlSurfaceShowcaseProps = {
  eyebrow?: string;
  sectionHeading: string;
  sectionDescription?: string;
  sectionHeadingId?: string;
  frameHeadingId?: string;
} & Omit<PublicControlSurfaceFrameProps, "headingId">;

export function PublicControlSurfaceShowcase({
  eyebrow,
  sectionHeading,
  sectionDescription,
  sectionHeadingId,
  frameHeadingId,
  ...frameProps
}: PublicControlSurfaceShowcaseProps) {
  return (
    <div className="vyn-public-surface-showcase">
      <PublicSectionIntro
        eyebrow={eyebrow}
        heading={sectionHeading}
        description={sectionDescription}
        headingId={sectionHeadingId}
      />
      <PublicControlSurfaceFrame {...frameProps} headingId={frameHeadingId} variant="showcase" />
    </div>
  );
}
