import type gsap from "gsap";
import type { ScrollTrigger as ScrollTriggerPlugin } from "gsap/ScrollTrigger";
import { getPreviewMotionProfile, PREVIEW_MOTION } from "./previewMotionConfig";

type GsapInstance = typeof gsap;

function revealSection(gsapLib: GsapInstance, section: HTMLElement, offsetY?: number): void {
  const profile = getPreviewMotionProfile();
  const inner = section.querySelector(".vyn-public-section-band-inner");
  if (!inner) {
    return;
  }

  gsapLib.from(inner, {
    opacity: 0,
    y: offsetY ?? profile.sectionOffsetY,
    duration: PREVIEW_MOTION.section.duration,
    ease: "power2.out",
    scrollTrigger: {
      trigger: section,
      start: "top 86%",
      once: true,
    },
  });
}

function animateHeroBlock(gsapLib: GsapInstance, heroSection: HTMLElement): void {
  const profile = getPreviewMotionProfile();
  const heroParts = gsapLib.utils.toArray<HTMLElement>(
    ".vyn-public-signal-path-step, .vyn-public-eyebrow, h1, .vyn-public-hero-description, .vyn-public-cta-stack",
    heroSection
  );
  if (heroParts.length > 0) {
    gsapLib.from(heroParts, {
      opacity: 0,
      y: profile.heroOffsetY,
      duration: PREVIEW_MOTION.hero.duration,
      stagger: PREVIEW_MOTION.hero.stagger,
      ease: "power2.out",
    });
  }

  const hero = heroSection.querySelector(".vyn-public-hero");
  if (hero) {
    const accentLine = hero.querySelector<HTMLElement>(".vyn-public-hero-accent-line");
    if (accentLine) {
      gsapLib.fromTo(
        accentLine,
        { scaleX: 0.35, opacity: 0.5 },
        {
          scaleX: 1,
          opacity: 1,
          duration: 0.5,
          ease: "power2.out",
        }
      );
    }
  }

  const companion = heroSection.querySelector(".vyn-public-surface");
  if (companion) {
    gsapLib.from(companion, {
      opacity: 0,
      x: profile.companionOffsetX,
      y: profile.companionOffsetY,
      duration: PREVIEW_MOTION.companion.duration,
      ease: "power2.out",
      delay: profile.companionOffsetX > 0 ? 0.12 : 0.06,
    });
  }
}

function animatePillarSection(gsapLib: GsapInstance, section: HTMLElement): void {
  const profile = getPreviewMotionProfile();
  const introTargets = gsapLib.utils.toArray<HTMLElement>(
    ".vyn-public-eyebrow, .vyn-public-section-title, .vyn-public-pillar-intro",
    section
  );
  const pillars = gsapLib.utils.toArray<HTMLElement>(".vyn-public-pillar", section);
  const glyphs = gsapLib.utils.toArray<HTMLElement>(".vyn-public-pillar-glyph-wrap", section);

  if (introTargets.length > 0) {
    gsapLib.from(introTargets, {
      opacity: 0,
      y: PREVIEW_MOTION.intro.offsetY,
      duration: PREVIEW_MOTION.intro.duration,
      stagger: 0.06,
      ease: "power2.out",
      scrollTrigger: {
        trigger: section,
        start: "top 86%",
        once: true,
      },
    });
  }

  if (pillars.length > 0) {
    gsapLib.from(pillars, {
      opacity: 0,
      y: profile.pillarOffsetY,
      duration: PREVIEW_MOTION.pillar.duration,
      stagger: PREVIEW_MOTION.pillar.stagger,
      ease: "power2.out",
      scrollTrigger: {
        trigger: section,
        start: "top 82%",
        once: true,
      },
    });
  }

  if (glyphs.length > 0) {
    gsapLib.from(glyphs, {
      opacity: 0,
      scale: 0.92,
      duration: 0.45,
      stagger: 0.08,
      ease: "power2.out",
      scrollTrigger: {
        trigger: section,
        start: "top 82%",
        once: true,
      },
    });
  }
}

function animateShowcaseSection(gsapLib: GsapInstance, section: HTMLElement): void {
  const profile = getPreviewMotionProfile();
  const intro = section.querySelector(".vyn-public-section-intro");
  const frame = section.querySelector(".vyn-public-surface-showcase-frame, .vyn-public-surface");

  if (intro) {
    gsapLib.from(intro, {
      opacity: 0,
      y: PREVIEW_MOTION.intro.offsetY,
      duration: PREVIEW_MOTION.intro.duration,
      ease: "power2.out",
      scrollTrigger: {
        trigger: section,
        start: "top 86%",
        once: true,
      },
    });
  }

  if (frame) {
    gsapLib.from(frame, {
      opacity: 0,
      y: profile.sectionOffsetY,
      duration: PREVIEW_MOTION.section.duration,
      ease: "power2.out",
      delay: 0.08,
      scrollTrigger: {
        trigger: section,
        start: "top 84%",
        once: true,
      },
    });
  }
}

export function bindPreviewHeaderScroll(
  _gsapLib: GsapInstance,
  scope: HTMLElement,
  scrollTrigger: typeof ScrollTriggerPlugin
): void {
  const shell = scope.closest(".vyn-public-shell");
  const header = shell?.querySelector<HTMLElement>(".vyn-public-header");
  if (!header) {
    return;
  }

  scrollTrigger.create({
    start: PREVIEW_MOTION.headerScrollEnd,
    onEnter: () => header.classList.add("vyn-public-header-scrolled"),
    onLeaveBack: () => header.classList.remove("vyn-public-header-scrolled"),
  });
}

export function applyDefaultPreviewMotion(gsapLib: GsapInstance, scope: HTMLElement): void {
  const profile = getPreviewMotionProfile();
  const sections = gsapLib.utils.toArray<HTMLElement>(".vyn-public-section-band", scope);
  if (sections.length === 0) {
    return;
  }

  const [heroSection, ...scrollSections] = sections;
  const heroInner = heroSection.querySelector(".vyn-public-section-band-inner");

  if (heroInner) {
    const heroTargets = gsapLib.utils.toArray<HTMLElement>(heroInner.children);
    gsapLib.from(heroTargets, {
      opacity: 0,
      y: Math.min(16, profile.heroOffsetY),
      duration: 0.5,
      stagger: 0.07,
      ease: "power2.out",
    });
  }

  scrollSections.forEach((section) => revealSection(gsapLib, section));
}

export function applyLandingPreviewMotion(gsapLib: GsapInstance, scope: HTMLElement): void {
  const sections = gsapLib.utils.toArray<HTMLElement>(".vyn-public-section-band", scope);
  if (sections.length === 0) {
    return;
  }

  const [heroSection, ...scrollSections] = sections;
  animateHeroBlock(gsapLib, heroSection);

  scrollSections.forEach((section) => {
    if (section.querySelector(".vyn-public-pillars")) {
      animatePillarSection(gsapLib, section);
      return;
    }

    if (section.querySelector(".vyn-public-surface-showcase")) {
      animateShowcaseSection(gsapLib, section);
      return;
    }

    revealSection(gsapLib, section);
  });
}
