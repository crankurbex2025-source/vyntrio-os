import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { LocaleSwitcher } from "../../../shared/ui/LocaleSwitcher";
import "./public-surface.css";

export type PublicFooterLink = {
  label: string;
  to: string;
};

export type PublicPreviewShellProps = {
  banner: string;
  brand: string;
  brandTo?: string;
  downloadLabel: string;
  downloadTo?: string;
  signInLabel: string;
  signInTo?: string;
  footerTagline: string;
  footerNote?: string;
  footerLinks?: PublicFooterLink[];
  routeLinks?: PublicFooterLink[];
  children: ReactNode;
  navAriaLabel?: string;
  anchorLinks?: PublicFooterLink[];
};

export function PublicPreviewShell({
  banner,
  brand,
  brandTo,
  downloadLabel,
  downloadTo = "/download",
  signInLabel,
  signInTo = "/login",
  footerTagline,
  footerNote,
  footerLinks,
  routeLinks,
  children,
  navAriaLabel = "Public navigation",
  anchorLinks,
}: PublicPreviewShellProps) {
  return (
    <div className="vyn-public-shell">
      <p className="vyn-public-banner" role="status">
        {banner}
      </p>

      <header className="vyn-public-header">
        <div className="vyn-public-header-inner">
          {brandTo ? (
            <Link className="vyn-public-brand" to={brandTo}>
              {brand}
            </Link>
          ) : (
            <span className="vyn-public-brand">{brand}</span>
          )}
          <nav className="vyn-public-nav" aria-label={navAriaLabel}>
            {routeLinks?.map((link) => (
              <Link key={link.to} to={link.to}>
                {link.label}
              </Link>
            ))}
            {anchorLinks?.map((link) => (
              <a key={link.to} href={link.to}>
                {link.label}
              </a>
            ))}
            <Link to={downloadTo}>{downloadLabel}</Link>
            <Link to={signInTo}>{signInLabel}</Link>
            <LocaleSwitcher />
          </nav>
        </div>
      </header>

      <main className="vyn-public-main">{children}</main>

      <footer className="vyn-public-footer">
        <div className="vyn-public-footer-inner">
          <div className="vyn-public-footer-top">
            <p className="vyn-public-footer-tagline">{footerTagline}</p>
            {footerLinks ? (
              <nav className="vyn-public-footer-nav" aria-label="Footer">
                {footerLinks.map((link) => (
                  <Link key={link.to} to={link.to}>
                    {link.label}
                  </Link>
                ))}
              </nav>
            ) : null}
          </div>
          {footerNote ? <p className="vyn-public-footer-note">{footerNote}</p> : null}
        </div>
      </footer>
    </div>
  );
}
