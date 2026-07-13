import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import "./App.css";
import { LoginScreen } from "./features/auth/LoginScreen";
import { SettingsShell } from "./features/settings/SettingsShell";
import { parsePublicSettingsDto, type PublicSettingsDto } from "./features/settings/settingsDto";
import { createApiClient, type ApiClient } from "./lib/api";

type AppState =
  | { kind: "bootstrapping" }
  | { kind: "unauthenticated" }
  | { kind: "probing_after_login"; csrfToken: string }
  | { kind: "authorized"; settings: PublicSettingsDto; csrfToken: string | null }
  | { kind: "forbidden" }
  | { kind: "unavailable" };

type AppProps = {
  apiClient?: ApiClient;
};

function SessionProbeView() {
  return (
    <main className="status-wrap">
      <section className="status-card">
        <h1>Vyntrio OS</h1>
        <p>Checking session...</p>
      </section>
    </main>
  );
}

function ForbiddenView() {
  return (
    <main className="status-wrap">
      <section className="status-card">
        <h1>Vyntrio OS</h1>
        <p>You do not have access to instance settings.</p>
      </section>
    </main>
  );
}

function UnavailableView() {
  return (
    <main className="status-wrap">
      <section className="status-card">
        <h1>Vyntrio OS</h1>
        <p>Instance settings are temporarily unavailable.</p>
      </section>
    </main>
  );
}

export default function App({ apiClient }: AppProps) {
  const client = useMemo(() => apiClient ?? createApiClient(), [apiClient]);
  const isMounted = useRef(true);
  const probeSequence = useRef(0);
  const [appState, setAppState] = useState<AppState>({ kind: "bootstrapping" });

  useEffect(() => {
    return () => {
      isMounted.current = false;
    };
  }, []);

  const probeSettings = useCallback(
    async (source: "initial" | "post_login", csrfToken: string | null) => {
      const sequence = ++probeSequence.current;
      const result = await client.requestJson<unknown>("/api/v1/settings");
      if (!isMounted.current || sequence !== probeSequence.current) {
        return;
      }

      if (!result.ok) {
        if (result.error.kind === "unauthorized") {
          setAppState({ kind: "unauthenticated" });
          return;
        }
        if (result.error.kind === "forbidden") {
          setAppState({ kind: "forbidden" });
          return;
        }
        setAppState({ kind: "unavailable" });
        return;
      }

      const settings = parsePublicSettingsDto(result.data);
      if (!settings) {
        setAppState({ kind: "unavailable" });
        return;
      }

      setAppState({
        kind: "authorized",
        settings,
        csrfToken: source === "post_login" ? csrfToken : null,
      });
    },
    [client]
  );

  useEffect(() => {
    void probeSettings("initial", null);
  }, [probeSettings]);

  const handleLoginSuccess = useCallback(
    (csrfToken: string) => {
      setAppState({ kind: "probing_after_login", csrfToken });
      void probeSettings("post_login", csrfToken);
    },
    [probeSettings]
  );

  switch (appState.kind) {
    case "bootstrapping":
    case "probing_after_login":
      return <SessionProbeView />;
    case "unauthenticated":
      return <LoginScreen apiClient={client} onLoginSuccess={handleLoginSuccess} />;
    case "forbidden":
      return <ForbiddenView />;
    case "unavailable":
      return <UnavailableView />;
    case "authorized":
      return <SettingsShell settings={appState.settings} />;
    default:
      return <UnavailableView />;
  }
}
