import { useCallback, useEffect, useId, useRef, useState, type ReactNode } from "react";
import { Link, NavLink, useLocation } from "react-router-dom";
import {
  applianceSectionFromPath,
  buildApplianceNavItems,
} from "./nav/applianceNavConfig";
import "./appliance-chrome.css";
import "./ui/appliance-ops.css";

type ApplianceShellProps = {
  children: ReactNode;
  /** When false, only wraps children (login/status screens). */
  withNav?: boolean;
  instanceName?: string;
  isSigningOut?: boolean;
  /** Disables sign-out without changing the button label (e.g. while saving settings). */
  signOutDisabled?: boolean;
  /** Override path-derived section (design-preview harness). */
  forceActiveId?: string;
  onSignOut?: () => void;
};

export function ApplianceShell({
  children,
  withNav = false,
  instanceName,
  isSigningOut = false,
  signOutDisabled = false,
  forceActiveId,
  onSignOut,
}: ApplianceShellProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const location = useLocation();
  const panelId = useId();
  const burgerRef = useRef<HTMLButtonElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);
  const navItems = buildApplianceNavItems();
  const activeId = forceActiveId ?? applianceSectionFromPath(location.pathname);

  const closeMenu = useCallback(() => setMenuOpen(false), []);

  useEffect(() => {
    closeMenu();
  }, [location.pathname, closeMenu]);

  useEffect(() => {
    if (!menuOpen) {
      document.body.style.removeProperty("overflow");
      return;
    }
    document.body.style.overflow = "hidden";
    const focusables = panelRef.current?.querySelectorAll<HTMLElement>(
      'a[href], button:not([disabled])'
    );
    focusables?.[0]?.focus();

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        closeMenu();
        burgerRef.current?.focus();
      }
    };
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("keydown", onKeyDown);
      document.body.style.removeProperty("overflow");
    };
  }, [menuOpen, closeMenu]);

  if (!withNav) {
    return <div className="vyn-appliance-shell">{children}</div>;
  }

  const navLinks = (
    <nav className="vyn-appliance-nav" aria-label="Appliance">
      {navItems.map((item) => (
        <NavLink
          key={item.id}
          to={item.to}
          end={item.to === "/app"}
          className={({ isActive }) => (isActive || activeId === item.id ? "active" : undefined)}
          onClick={closeMenu}
        >
          <span>{item.label}</span>
          {item.planned ? <span className="vyn-appliance-nav-planned">Planned</span> : null}
        </NavLink>
      ))}
    </nav>
  );

  return (
    <div className="vyn-appliance-shell vyn-appliance-shell-ops">
      <div className="vyn-appliance-frame">
        <header className="vyn-appliance-topbar">
          <Link className="vyn-appliance-brand" to="/app" onClick={closeMenu}>
            <span className="vyn-appliance-brand-mark" aria-hidden="true" />
            {instanceName?.trim() || "Vyntrio"}
          </Link>
          <div className="vyn-appliance-topbar-actions">
            {onSignOut ? (
              <button
                type="button"
                className="vyn-appliance-signout"
                onClick={onSignOut}
                disabled={isSigningOut || signOutDisabled}
              >
                {isSigningOut ? "Signing out..." : "Sign out"}
              </button>
            ) : null}
            <button
              ref={burgerRef}
              type="button"
              className="vyn-appliance-burger"
              aria-expanded={menuOpen}
              aria-controls={panelId}
              aria-label={menuOpen ? "Close menu" : "Open menu"}
              onClick={() => setMenuOpen((open) => !open)}
            >
              <span className="vyn-appliance-burger-lines" aria-hidden="true">
                <span />
                <span />
                <span />
              </span>
            </button>
          </div>
        </header>

        <aside className="vyn-appliance-sidebar" aria-label="Sections">
          {navLinks}
        </aside>

        {menuOpen ? (
          <button
            type="button"
            className="vyn-appliance-scrim"
            aria-label="Close menu"
            onClick={closeMenu}
          />
        ) : null}

        <div
          ref={panelRef}
          id={panelId}
          className="vyn-appliance-drawer"
          hidden={!menuOpen}
          role="dialog"
          aria-modal="true"
          aria-label="Appliance menu"
        >
          <div className="vyn-appliance-drawer-header">
            <span>Menu</span>
            <button
              type="button"
              className="vyn-appliance-drawer-close"
              aria-label="Close menu"
              onClick={closeMenu}
            >
              ✕
            </button>
          </div>
          {navLinks}
        </div>

        <div className="vyn-appliance-content">{children}</div>
      </div>
    </div>
  );
}
