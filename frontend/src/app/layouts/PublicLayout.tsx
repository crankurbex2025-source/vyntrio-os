import { Link, Outlet } from "react-router-dom";
import "../../surfaces/public/public.css";

export function PublicLayout() {
  return (
    <div className="public-shell">
      <header className="public-header">
        <div className="public-header-inner">
          <Link className="public-brand" to="/">
            <span className="public-brand-mark" aria-hidden="true" />
            <span className="public-brand-text">Vyntrio</span>
          </Link>
          <nav className="public-nav" aria-label="Public site">
            <Link className="public-nav-link" to="/">
              Home
            </Link>
            <Link className="public-nav-link" to="/download">
              Download
            </Link>
            <Link className="public-nav-link public-nav-link-accent" to="/login">
              Sign in
            </Link>
          </nav>
        </div>
      </header>
      <Outlet />
      <footer className="public-footer">
        <div className="public-footer-inner">
          <p>Vyntrio OS — local appliance platform for storage, services, and control.</p>
        </div>
      </footer>
    </div>
  );
}
