import type { PublicMessages } from "../../../shared/i18n";
import type { PublicPreviewShellProps } from "../components/PublicPreviewShell";
import { buildProductionNav } from "../nav/publicNavConfig";

type ProductionShellOptions = Pick<
  PublicPreviewShellProps,
  "navAriaLabel" | "shellVariant"
>;

export function buildProductionShellProps(
  messages: PublicMessages,
  options: ProductionShellOptions = {}
): Omit<PublicPreviewShellProps, "children"> {
  return {
    nav: buildProductionNav(messages),
    navAriaLabel: options.navAriaLabel ?? messages.chrome.navAriaLabel,
    menuOpenLabel: messages.chrome.menuOpen,
    menuCloseLabel: messages.chrome.menuClose,
    mobileNavLabel: messages.chrome.mobileNav,
    themeToggleLabel: messages.chrome.themeToggle,
    premiumSurface: true,
    shellVariant: options.shellVariant ?? "landing",
  };
}
