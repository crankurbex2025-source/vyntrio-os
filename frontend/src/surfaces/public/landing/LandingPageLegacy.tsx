import { Link } from "react-router-dom";
import { PageHeader } from "../../../shared/ui/PageHeader";

/** Slice 11.1 landing — retained at /design-preview/landing-legacy for rollback review. */
export function LandingPageLegacy() {
  return (
    <main className="public-main">
      <section className="landing-hero">
        <div className="landing-hero-panel">
          <PageHeader
            eyebrow="Appliance platform"
            title="Your server, under your control."
            description="Vyntrio OS is a local appliance platform for storage, services, and administration — designed to run on your hardware, not in someone else's cloud."
          />
          <div className="landing-hero-actions">
            <Link className="public-button public-button-primary" to="/download">
              Get Vyntrio OS
            </Link>
            <Link className="public-button public-button-secondary" to="/login">
              Sign in to your appliance
            </Link>
          </div>
        </div>

        <section aria-labelledby="landing-value-heading">
          <h2 id="landing-value-heading" className="landing-section-title">
            Built for operators, not subscriptions
          </h2>
          <div className="landing-value-grid">
            <article className="landing-value-card">
              <h2>Local-first control plane</h2>
              <p>
                Manage storage, services, and system state from a control surface that runs on the
                appliance itself — same origin, same trust boundary.
              </p>
            </article>
            <article className="landing-value-card">
              <h2>Appliance clarity</h2>
              <p>
                A focused overview for system health, runtime readiness, and instance identity —
                structured for day-to-day server operations, not dashboard theater.
              </p>
            </article>
            <article className="landing-value-card">
              <h2>Honest product surface</h2>
              <p>
                Public pages stay static and truthful. Live status and admin actions appear only
                after you sign in to a real appliance.
              </p>
            </article>
          </div>
        </section>
      </section>
    </main>
  );
}
