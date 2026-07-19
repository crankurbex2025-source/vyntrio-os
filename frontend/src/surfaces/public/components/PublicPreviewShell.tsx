import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import type { PublicNavModel } from "../nav/publicNavConfig";
import { PublicSiteHeader } from "./PublicSiteHeader";
import { usePublicTheme } from "./usePublicTheme";
import "./public-surface.css";
import "./public-responsive.css";
import "./public-chrome.css";

export type PublicFooterLink = {
  label: string;
  to: string;
};

/** @deprecated Prefer `nav` model — kept for transitional callers/tests. */
export type PublicPreviewShellProps = {
  banner?: string;
  brand?: string;
  brandTo?: string;
  downloadLabel?: string;
  downloadTo?: string;
  signInLabel?: string;
  signInTo?: string;
  footerTagline?: string;
  footerNote?: string;
  footerLinks?: PublicFooterLink[];
  routeLinks?: PublicFooterLink[];
  children: ReactNode;
  navAriaLabel?: string;
  /** Ignored — nav duplication removed. Kept so old call sites typecheck during migration. */
  anchorLinks?: PublicFooterLink[];
  premiumSurface?: boolean;
  shellVariant?: "default" | "landing" | "product";
  /** Canonical navigation model (preferred). */
  nav?: PublicNavModel;
  menuOpenLabel?: string;
  menuCloseLabel?: string;
  mobileNavLabel?: string;
  themeToggleLabel?: string;
};

function legacyNavFromProps(props: PublicPreviewShellProps): PublicNavModel | null {
  if (!props.brand || !props.downloadLabel || !props.signInLabel || !props.footerTagline) {
    return null;
  }
  const items = (props.routeLinks ?? [])
    .filter((link) => link.to !== props.downloadTo)
    .map((link, index) => ({
      id: `legacy-${index}`,
      label: link.label,
      to: link.to,
    }));
  return {
    brand: props.brand,
    brandTo: props.brandTo ?? "/",
    items,
    cta: {
      id: "download",
      label: props.downloadLabel,
      to: props.downloadTo ?? "/download",
      cta: true,
    },
    secondary: {
      id: "sign-in",
      label: props.signInLabel,
      to: props.signInTo ?? "/login",
      secondary: true,
    },
    footerItems: (props.footerLinks ?? []).map((link, index) => ({
      id: `footer-${index}`,
      label: link.label,
      to: link.to,
    })),
    footerTagline: props.footerTagline,
    footerNote: props.footerNote,
  };
}

export function PublicPreviewShell({
  banner,
  children,
  navAriaLabel = "Primary",
  premiumSurface = true,
  shellVariant = "default",
  nav: navProp,
  menuOpenLabel = "Open menu",
  menuCloseLabel = "Close menu",
  mobileNavLabel = "Site menu",
  themeToggleLabel = "Toggle color theme",
  ...legacy
}: PublicPreviewShellProps) {
  const { theme, toggleTheme } = usePublicTheme();
  const nav = navProp ?? legacyNavFromProps({ ...legacy, children, navAriaLabel });

  if (!nav) {
    throw new Error("PublicPreviewShell requires a nav model or legacy brand/download props");
  }

  const shellClass = [
    "vyn-public-shell",
    premiumSurface ? "vyn-public-shell-premium" : "",
    shellVariant === "landing" ? "vyn-public-shell-landing" : "",
    shellVariant === "product" ? "vyn-public-shell-product" : "",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <div className={shellClass} data-theme={theme}>
      {banner ? (
        <p className="vyn-public-banner" role="status">
          {banner}
        </p>
      ) : null}

      <PublicSiteHeader
        nav={nav}
        navAriaLabel={navAriaLabel}
        theme={theme}
        onToggleTheme={toggleTheme}
        themeToggleLabel={themeToggleLabel}
        menuOpenLabel={menuOpenLabel}
        menuCloseLabel={menuCloseLabel}
        mobileNavLabel={mobileNavLabel}
      />

      <main className="vyn-public-main">{children}</main>

      <footer className="vyn-public-footer">
        <div className="vyn-public-footer-inner">
          <div className="vyn-public-footer-top">
            <p className="vyn-public-footer-tagline">{nav.footerTagline}</p>
            <nav className="vyn-public-footer-nav" aria-label="Footer">
              {nav.footerItems.map((link) =>
                link.to.includes("#") ? (
                  <a key={link.id} href={link.to}>
                    {link.label}
                  </a>
                ) : (
                  <Link key={link.id} to={link.to}>
                    {link.label}
                  </Link>
                )
              )}
            </nav>
          </div>
          {nav.footerNote ? <p className="vyn-public-footer-note">{nav.footerNote}</p> : null}
        </div>
      </footer>
    </div>
  );
}
