import type { PublicMessages } from "../../../shared/i18n";
import type { PublicPreviewShellProps } from "../components/PublicPreviewShell";
import { buildPreviewNav } from "../nav/publicNavConfig";

type PreviewShellOptions = Pick<PublicPreviewShellProps, "navAriaLabel">;

export function buildPreviewShellProps(
  messages: PublicMessages,
  options: PreviewShellOptions = {}
): Omit<PublicPreviewShellProps, "children"> {
  return {
    banner: messages.previewBanner,
    nav: buildPreviewNav(messages),
    navAriaLabel: options.navAriaLabel ?? messages.chrome.navAriaLabel,
    menuOpenLabel: messages.chrome.menuOpen,
    menuCloseLabel: messages.chrome.menuClose,
    mobileNavLabel: messages.chrome.mobileNav,
    themeToggleLabel: messages.chrome.themeToggle,
    premiumSurface: true,
  };
}
