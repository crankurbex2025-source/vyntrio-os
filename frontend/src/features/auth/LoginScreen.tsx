import { useEffect, useMemo, useRef, useState, type FormEvent } from "react";
import { createApiClient, type ApiClient } from "../../lib/api";

type LoginViewState =
  | { kind: "idle" }
  | { kind: "submitting" }
  | { kind: "error" }
  | { kind: "signed_in"; csrfToken: string };

type LoginScreenProps = {
  apiClient?: ApiClient;
};

const GENERIC_LOGIN_ERROR = "Sign-in failed. Check your credentials and try again.";

function readCSRFTokenFromLoginPayload(payload: unknown): string | null {
  if (typeof payload !== "object" || payload === null || Array.isArray(payload)) {
    return null;
  }

  const record = payload as Record<string, unknown>;
  const keys = Object.keys(record);
  if (keys.length !== 1 || keys[0] !== "csrf_token") {
    return null;
  }

  const csrfToken = record.csrf_token;
  if (typeof csrfToken !== "string" || csrfToken.length === 0) {
    return null;
  }

  return csrfToken;
}

export function LoginScreen({ apiClient }: LoginScreenProps) {
  const client = useMemo(() => apiClient ?? createApiClient(), [apiClient]);
  const isMounted = useRef(true);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [viewState, setViewState] = useState<LoginViewState>({ kind: "idle" });

  useEffect(() => {
    return () => {
      isMounted.current = false;
    };
  }, []);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (viewState.kind === "submitting") {
      return;
    }

    if (isMounted.current) {
      setViewState({ kind: "submitting" });
    }

    try {
      const result = await client.requestJson<unknown>("/api/v1/identity/login", {
        method: "POST",
        jsonBody: {
          username,
          password,
        },
      });

      if (!isMounted.current) {
        return;
      }

      if (!result.ok) {
        setViewState({ kind: "error" });
        return;
      }

      const csrfToken = readCSRFTokenFromLoginPayload(result.data);
      if (!csrfToken) {
        setViewState({ kind: "error" });
        return;
      }

      setViewState({ kind: "signed_in", csrfToken });
    } catch {
      if (isMounted.current) {
        setViewState({ kind: "error" });
      }
    } finally {
      if (isMounted.current) {
        setPassword("");
      }
    }
  }

  function handleReset() {
    setViewState({ kind: "idle" });
    setPassword("");
  }

  if (viewState.kind === "signed_in") {
    return (
      <section className="auth-card" aria-live="polite">
        <h1>Vyntrio OS</h1>
        <p>Sign-in succeeded for this browser session.</p>
        <button type="button" onClick={handleReset}>
          Reset sign-in view
        </button>
      </section>
    );
  }

  const isSubmitting = viewState.kind === "submitting";

  return (
    <main className="auth-wrap">
      <form className="auth-card" onSubmit={handleSubmit}>
        <h1>Vyntrio OS</h1>
        <p>Sign in</p>

        <label htmlFor="username">Username</label>
        <input
          id="username"
          name="username"
          type="text"
          autoComplete="username"
          value={username}
          onChange={(event) => setUsername(event.target.value)}
          disabled={isSubmitting}
          required
        />

        <label htmlFor="password">Password</label>
        <input
          id="password"
          name="password"
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          disabled={isSubmitting}
          required
        />

        {viewState.kind === "error" ? <p role="alert">{GENERIC_LOGIN_ERROR}</p> : null}

        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Signing in..." : "Sign in"}
        </button>
      </form>
    </main>
  );
}
