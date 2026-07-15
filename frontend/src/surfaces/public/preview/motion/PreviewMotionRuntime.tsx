import type { ReactNode } from "react";
import { useRef } from "react";
import gsap from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";
import { useGSAP } from "@gsap/react";
import {
  applyDefaultPreviewMotion,
  applyLandingPreviewMotion,
  bindPreviewHeaderScroll,
} from "./applyPreviewMotion";
import type { PreviewMotionVariant } from "./previewMotionConfig";

let pluginsRegistered = false;

function registerPreviewMotionPlugins(): void {
  if (pluginsRegistered) {
    return;
  }

  gsap.registerPlugin(ScrollTrigger);
  pluginsRegistered = true;
}

export type PreviewMotionRuntimeProps = {
  children: ReactNode;
  variant?: PreviewMotionVariant;
};

export function PreviewMotionRuntime({ children, variant = "default" }: PreviewMotionRuntimeProps) {
  const scope = useRef<HTMLDivElement>(null);

  useGSAP(
    () => {
      if (!scope.current) {
        return;
      }

      registerPreviewMotionPlugins();
      bindPreviewHeaderScroll(gsap, scope.current, ScrollTrigger);

      if (variant === "landing") {
        applyLandingPreviewMotion(gsap, scope.current);
        return;
      }

      applyDefaultPreviewMotion(gsap, scope.current);
    },
    { scope, dependencies: [variant] }
  );

  return (
    <div
      ref={scope}
      className="vyn-preview-motion-scope"
      data-motion="on"
      data-motion-variant={variant}
    >
      {children}
    </div>
  );
}
