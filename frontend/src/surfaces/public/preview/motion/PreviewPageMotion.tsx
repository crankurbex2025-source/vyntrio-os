import type { ReactNode } from "react";
import { useRef } from "react";
import gsap from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";
import { useGSAP } from "@gsap/react";
import { prefersReducedMotion } from "./prefersReducedMotion";

const HERO_STAGGER = 0.07;
const HERO_DURATION = 0.5;
const SECTION_DURATION = 0.6;
const SECTION_OFFSET = 22;

let pluginsRegistered = false;

function registerPreviewMotionPlugins(): void {
  if (pluginsRegistered) {
    return;
  }

  gsap.registerPlugin(ScrollTrigger);
  pluginsRegistered = true;
}

export type PreviewPageMotionProps = {
  children: ReactNode;
};

export function PreviewPageMotion({ children }: PreviewPageMotionProps) {
  const scope = useRef<HTMLDivElement>(null);
  const reducedMotion = prefersReducedMotion();

  useGSAP(
    () => {
      if (reducedMotion || !scope.current) {
        return;
      }

      registerPreviewMotionPlugins();

      const sections = gsap.utils.toArray<HTMLElement>(
        ".vyn-public-section-band",
        scope.current
      );

      if (sections.length === 0) {
        return;
      }

      const [heroSection, ...scrollSections] = sections;
      const heroInner = heroSection.querySelector(".vyn-public-section-band-inner");

      if (heroInner) {
        const heroTargets = gsap.utils.toArray<HTMLElement>(heroInner.children);
        gsap.from(heroTargets, {
          opacity: 0,
          y: 16,
          duration: HERO_DURATION,
          stagger: HERO_STAGGER,
          ease: "power2.out",
        });
      }

      scrollSections.forEach((section) => {
        const inner = section.querySelector(".vyn-public-section-band-inner");
        if (!inner) {
          return;
        }

        gsap.from(inner, {
          opacity: 0,
          y: SECTION_OFFSET,
          duration: SECTION_DURATION,
          ease: "power2.out",
          scrollTrigger: {
            trigger: section,
            start: "top 88%",
            once: true,
          },
        });
      });
    },
    { scope, dependencies: [reducedMotion] }
  );

  return (
    <div ref={scope} className="vyn-preview-motion-scope" data-motion={reducedMotion ? "off" : "on"}>
      {children}
    </div>
  );
}
