import { useCallback, useEffect, useId, useRef, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { LocaleSwitcher } from "../../../shared/ui/LocaleSwitcher";
import type { PublicNavModel } from "../nav/publicNavConfig";
import type { PublicTheme } from "./usePublicTheme";

export type PublicSiteHeaderProps = {
  nav: PublicNavModel;
  navAriaLabel?: string;
  theme: PublicTheme;
  onToggleTheme: () => void;
  themeToggleLabel: string;
  menuOpenLabel: string;
  menuCloseLabel: string;
  mobileNavLabel: string;
};

function NavLinkItem({
  to,
  label,
  onNavigate,
  className,
}: {
  to: string;
  label: string;
  onNavigate?: () => void;
  className?: string;
}) {
  const isHash = to.includes("#");
  if (isHash && !to.startsWith("/design-preview")) {
    // Hash links on production: use <a> so in-page anchors work from any route.
    return (
      <a className={className} href={to} onClick={onNavigate}>
        {label}
      </a>
    );
  }
  if (to.includes("#") && to.startsWith("/design-preview")) {
    return (
      <a className={className} href={to} onClick={onNavigate}>
        {label}
      </a>
    );
  }
  return (
    <Link className={className} to={to} onClick={onNavigate}>
      {label}
    </Link>
  );
}

export function PublicSiteHeader({
  nav,
  navAriaLabel = "Primary",
  theme,
  onToggleTheme,
  themeToggleLabel,
  menuOpenLabel,
  menuCloseLabel,
  mobileNavLabel,
}: PublicSiteHeaderProps) {
  const [open, setOpen] = useState(false);
  const [scrolled, setScrolled] = useState(false);
  const location = useLocation();
  const panelId = useId();
  const burgerRef = useRef<HTMLButtonElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);
  const previouslyFocused = useRef<HTMLElement | null>(null);

  const closeMenu = useCallback(() => {
    setOpen(false);
  }, []);

  useEffect(() => {
    closeMenu();
  }, [location.pathname, location.hash, closeMenu]);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  useEffect(() => {
    if (!open) {
      document.body.style.removeProperty("overflow");
      if (previouslyFocused.current) {
        previouslyFocused.current.focus();
        previouslyFocused.current = null;
      }
      return;
    }

    previouslyFocused.current = document.activeElement as HTMLElement | null;
    document.body.style.overflow = "hidden";

    const panel = panelRef.current;
    const focusables = panel?.querySelectorAll<HTMLElement>(
      'a[href], button:not([disabled]), [tabindex]:not([tabindex="-1"])'
    );
    focusables?.[0]?.focus();

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        closeMenu();
        burgerRef.current?.focus();
        return;
      }
      if (event.key !== "Tab" || !focusables || focusables.length === 0) {
        return;
      }
      const list = Array.from(focusables);
      const first = list[0];
      const last = list[list.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    };

    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("keydown", onKeyDown);
      document.body.style.removeProperty("overflow");
    };
  }, [open, closeMenu]);

  const headerClass = [
    "vyn-public-header",
    scrolled ? "vyn-public-header-scrolled" : "",
    open ? "vyn-public-header-menu-open" : "",
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <header className={headerClass}>
      <div className="vyn-public-header-inner">
        <Link className="vyn-public-brand" to={nav.brandTo} onClick={closeMenu}>
          <span className="vyn-public-brand-mark" aria-hidden="true" />
          {nav.brand}
        </Link>

        <nav className="vyn-public-nav vyn-public-nav-desktop" aria-label={navAriaLabel}>
          {nav.items.map((item) => (
            <NavLinkItem key={item.id} to={item.to} label={item.label} />
          ))}
        </nav>

        <div className="vyn-public-header-actions">
          <button
            type="button"
            className="vyn-public-theme-toggle"
            onClick={onToggleTheme}
            aria-label={themeToggleLabel}
            title={themeToggleLabel}
          >
            <span className="vyn-public-theme-icon" aria-hidden="true">
              {theme === "light" ? (
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                  <path
                    d="M21 14.5A8.5 8.5 0 0 1 9.5 3 7 7 0 1 0 21 14.5Z"
                    stroke="currentColor"
                    strokeWidth="1.75"
                    strokeLinejoin="round"
                  />
                </svg>
              ) : (
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="4" stroke="currentColor" strokeWidth="1.75" />
                  <path
                    d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"
                    stroke="currentColor"
                    strokeWidth="1.75"
                    strokeLinecap="round"
                  />
                </svg>
              )}
            </span>
          </button>
          <LocaleSwitcher />
          <NavLinkItem
            className="vyn-public-nav-secondary vyn-public-nav-desktop-only"
            to={nav.secondary.to}
            label={nav.secondary.label}
          />
          <NavLinkItem
            className="vyn-public-nav-cta vyn-public-nav-desktop-only"
            to={nav.cta.to}
            label={nav.cta.label}
          />
          <button
            ref={burgerRef}
            type="button"
            className="vyn-public-burger"
            aria-expanded={open}
            aria-controls={panelId}
            aria-label={open ? menuCloseLabel : menuOpenLabel}
            onClick={() => setOpen((value) => !value)}
          >
            <span className="vyn-public-burger-lines" aria-hidden="true">
              <span />
              <span />
              <span />
            </span>
          </button>
        </div>
      </div>

      <div
        className={`vyn-public-nav-scrim${open ? " is-open" : ""}`}
        hidden={!open}
        onClick={closeMenu}
      />

      <div
        ref={panelRef}
        id={panelId}
        className={`vyn-public-nav-drawer${open ? " is-open" : ""}`}
        role="dialog"
        aria-modal="true"
        aria-label={mobileNavLabel}
        hidden={!open}
      >
        <nav className="vyn-public-nav-mobile" aria-label={mobileNavLabel}>
          {nav.items.map((item) => (
            <NavLinkItem
              key={item.id}
              to={item.to}
              label={item.label}
              onNavigate={closeMenu}
            />
          ))}
          <NavLinkItem
            className="vyn-public-nav-secondary"
            to={nav.secondary.to}
            label={nav.secondary.label}
            onNavigate={closeMenu}
          />
          <NavLinkItem
            className="vyn-public-nav-cta"
            to={nav.cta.to}
            label={nav.cta.label}
            onNavigate={closeMenu}
          />
        </nav>
      </div>
    </header>
  );
}
