export const PREVIEW_MOTION = {
  hero: { duration: 0.62, stagger: 0.09, offsetY: 18 },
  companion: { duration: 0.72, offsetX: 28, offsetY: 10 },
  section: { duration: 0.68, offsetY: 26 },
  pillar: { duration: 0.55, stagger: 0.1, offsetY: 20 },
  intro: { duration: 0.58, offsetY: 16 },
  headerScrollEnd: 140,
  narrow: {
    hero: { offsetY: 12 },
    companion: { offsetX: 0, offsetY: 8 },
    section: { offsetY: 18 },
    pillar: { offsetY: 14 },
  },
} as const;

export type PreviewMotionVariant = "default" | "landing";

export type PreviewMotionProfile = {
  heroOffsetY: number;
  companionOffsetX: number;
  companionOffsetY: number;
  sectionOffsetY: number;
  pillarOffsetY: number;
};

const DEFAULT_PROFILE: PreviewMotionProfile = {
  heroOffsetY: PREVIEW_MOTION.hero.offsetY,
  companionOffsetX: PREVIEW_MOTION.companion.offsetX,
  companionOffsetY: PREVIEW_MOTION.companion.offsetY,
  sectionOffsetY: PREVIEW_MOTION.section.offsetY,
  pillarOffsetY: PREVIEW_MOTION.pillar.offsetY,
};

const NARROW_PROFILE: PreviewMotionProfile = {
  heroOffsetY: PREVIEW_MOTION.narrow.hero.offsetY,
  companionOffsetX: PREVIEW_MOTION.narrow.companion.offsetX,
  companionOffsetY: PREVIEW_MOTION.narrow.companion.offsetY,
  sectionOffsetY: PREVIEW_MOTION.narrow.section.offsetY,
  pillarOffsetY: PREVIEW_MOTION.narrow.pillar.offsetY,
};

export function getPreviewMotionProfile(): PreviewMotionProfile {
  if (typeof window === "undefined" || typeof window.matchMedia !== "function") {
    return DEFAULT_PROFILE;
  }

  return window.matchMedia("(max-width: 40rem)").matches ? NARROW_PROFILE : DEFAULT_PROFILE;
}
