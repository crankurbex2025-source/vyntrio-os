import { lazy, Suspense, type ReactNode } from "react";
import type { PreviewMotionVariant } from "./previewMotionConfig";
import { prefersReducedMotion } from "./prefersReducedMotion";

const PreviewMotionRuntime = lazy(() =>
  import("./PreviewMotionRuntime").then((module) => ({ default: module.PreviewMotionRuntime }))
);

export type PreviewPageMotionProps = {
  children: ReactNode;
  variant?: PreviewMotionVariant;
};

function PreviewMotionPending({
  children,
  variant,
}: {
  children: ReactNode;
  variant: PreviewMotionVariant;
}) {
  return (
    <div
      className="vyn-preview-motion-scope"
      data-motion="pending"
      data-motion-variant={variant}
    >
      {children}
    </div>
  );
}

export function PreviewPageMotion({ children, variant = "default" }: PreviewPageMotionProps) {
  const reducedMotion = prefersReducedMotion();

  if (reducedMotion) {
    return (
      <div
        className="vyn-preview-motion-scope"
        data-motion="off"
        data-motion-variant={variant}
      >
        {children}
      </div>
    );
  }

  return (
    <Suspense fallback={<PreviewMotionPending variant={variant}>{children}</PreviewMotionPending>}>
      <PreviewMotionRuntime variant={variant}>{children}</PreviewMotionRuntime>
    </Suspense>
  );
}
