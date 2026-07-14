import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import "./App.css";
import { LoginScreen } from "./features/auth/LoginScreen";
import { OverviewShell } from "./features/overview/OverviewShell";
import { parseOverviewDto, type OverviewDto } from "./features/overview/overviewDto";
import { SettingsShell } from "./features/settings/SettingsShell";
import { parsePublicSettingsDto, type PublicSettingsDto } from "./features/settings/settingsDto";
import { createApiClient, type ApiClient } from "./lib/api";

const UPDATE_INSTANCE_ERROR_MESSAGE = "The instance name could not be updated. Please try again.";
const UPDATE_INSTANCE_VALIDATION_MESSAGE = "Enter a valid instance name.";

type AuthorizedState = {
  kind: "authorized";
  overview: OverviewDto;
  csrfToken: string | null;
  view: "overview" | "settings";
  settings: PublicSettingsDto | null;
  logoutPending: boolean;
  logoutError: boolean;
  settingsLoading: boolean;
  settingsAccessError: boolean;
  editMode: boolean;
  draftDisplayName: string;
  updatePending: boolean;
  updateError: boolean;
  updateValidationError: boolean;
};

type AppState =
  | { kind: "bootstrapping" }
  | { kind: "unauthenticated" }
  | { kind: "probing_after_login"; csrfToken: string }
  | AuthorizedState
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
        <p>You do not have access to this appliance view.</p>
      </section>
    </main>
  );
}

function UnavailableView() {
  return (
    <main className="status-wrap">
      <section className="status-card">
        <h1>Vyntrio OS</h1>
        <p>The appliance overview is temporarily unavailable.</p>
      </section>
    </main>
  );
}

export default function App({ apiClient }: AppProps) {
  const client = useMemo(() => apiClient ?? createApiClient(), [apiClient]);
  const isMounted = useRef(true);
  const probeSequence = useRef(0);
  const logoutSequence = useRef(0);
  const updateSequence = useRef(0);
  const [appState, setAppState] = useState<AppState>({ kind: "bootstrapping" });

  const buildAuthorizedState = useCallback(
    (overview: OverviewDto, csrfToken: string | null): AuthorizedState => ({
      kind: "authorized",
      overview,
      csrfToken,
      view: "overview",
      settings: null,
      logoutPending: false,
      logoutError: false,
      settingsLoading: false,
      settingsAccessError: false,
      editMode: false,
      draftDisplayName: overview.instance.name,
      updatePending: false,
      updateError: false,
      updateValidationError: false,
    }),
    []
  );

  useEffect(() => {
    return () => {
      isMounted.current = false;
    };
  }, []);

  const probeOverview = useCallback(
    async (source: "initial" | "post_login", csrfToken: string | null) => {
      const sequence = ++probeSequence.current;
      const result = await client.requestJson<unknown>("/api/v1/overview");
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

      const overview = parseOverviewDto(result.data);
      if (!overview) {
        setAppState({ kind: "unavailable" });
        return;
      }

      setAppState(buildAuthorizedState(overview, source === "post_login" ? csrfToken : null));
    },
    [buildAuthorizedState, client]
  );

  useEffect(() => {
    void probeOverview("initial", null);
  }, [probeOverview]);

  const handleLoginSuccess = useCallback(
    (csrfToken: string) => {
      setAppState({ kind: "probing_after_login", csrfToken });
      void probeOverview("post_login", csrfToken);
    },
    [probeOverview]
  );

  const handleOpenSettings = useCallback(async () => {
    if (appState.kind !== "authorized") {
      return;
    }
    if (appState.logoutPending || appState.settingsLoading || appState.updatePending) {
      return;
    }

    setAppState({
      ...appState,
      settingsLoading: true,
      settingsAccessError: false,
    });

    const result = await client.requestJson<unknown>("/api/v1/settings");
    if (!isMounted.current) {
      return;
    }

    if (!result.ok) {
      if (result.error.kind === "unauthorized") {
        setAppState({ kind: "unauthenticated" });
        return;
      }
      if (result.error.kind === "forbidden") {
        setAppState((previous) => {
          if (previous.kind !== "authorized") {
            return previous;
          }
          return {
            ...previous,
            settingsLoading: false,
            settingsAccessError: true,
            view: "overview",
          };
        });
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

    setAppState((previous) => {
      if (previous.kind !== "authorized") {
        return previous;
      }
      return {
        ...previous,
        view: "settings",
        settings,
        settingsLoading: false,
        settingsAccessError: false,
        editMode: false,
        draftDisplayName: settings.instance.name,
        updatePending: false,
        updateError: false,
        updateValidationError: false,
      };
    });
  }, [appState, client]);

  const handleBackToOverview = useCallback(() => {
    setAppState((previous) => {
      if (previous.kind !== "authorized") {
        return previous;
      }
      return {
        ...previous,
        view: "overview",
        settings: null,
        editMode: false,
        draftDisplayName: previous.overview.instance.name,
        updatePending: false,
        updateError: false,
        updateValidationError: false,
      };
    });
  }, []);

  const handleSignOut = useCallback(async () => {
    if (appState.kind !== "authorized") {
      return;
    }
    if (appState.logoutPending || appState.updatePending) {
      return;
    }
    if (!appState.csrfToken) {
      setAppState({ ...appState, logoutPending: false, logoutError: true });
      return;
    }

    if (!client.requestNoContent) {
      setAppState({ ...appState, logoutPending: false, logoutError: true });
      return;
    }

    const csrfTokenToUse = appState.csrfToken;
    setAppState({ ...appState, logoutPending: true, logoutError: false });

    const sequence = ++logoutSequence.current;
    const result = await client.requestNoContent("/api/v1/identity/logout", 204, {
      method: "POST",
      csrfToken: csrfTokenToUse,
    });

    if (!isMounted.current || sequence !== logoutSequence.current) {
      return;
    }

    if (result.ok) {
      setAppState({ kind: "unauthenticated" });
      return;
    }

    setAppState((previous) => {
      if (previous.kind !== "authorized") {
        return previous;
      }
      return { ...previous, logoutPending: false, logoutError: true };
    });
  }, [appState, client]);

  const handleStartEdit = useCallback(() => {
    setAppState((previous) => {
      if (previous.kind !== "authorized" || previous.view !== "settings" || !previous.settings) {
        return previous;
      }
      if (previous.logoutPending || previous.updatePending) {
        return previous;
      }
      return {
        ...previous,
        editMode: true,
        draftDisplayName: previous.settings.instance.name,
        updateError: false,
        updateValidationError: false,
      };
    });
  }, []);

  const handleCancelEdit = useCallback(() => {
    setAppState((previous) => {
      if (previous.kind !== "authorized" || !previous.settings) {
        return previous;
      }
      if (previous.logoutPending || previous.updatePending) {
        return previous;
      }
      return {
        ...previous,
        editMode: false,
        draftDisplayName: previous.settings.instance.name,
        updateError: false,
        updateValidationError: false,
      };
    });
  }, []);

  const handleDraftDisplayNameChange = useCallback((value: string) => {
    setAppState((previous) => {
      if (previous.kind !== "authorized") {
        return previous;
      }
      return {
        ...previous,
        draftDisplayName: value,
        updateValidationError: false,
      };
    });
  }, []);

  const handleSaveDisplayName = useCallback(async () => {
    if (appState.kind !== "authorized" || appState.view !== "settings" || !appState.settings) {
      return;
    }
    if (!appState.editMode || appState.logoutPending || appState.updatePending) {
      return;
    }

    const normalizedDisplayName = validateInstanceDisplayNameDraft(appState.draftDisplayName);
    if (!normalizedDisplayName) {
      setAppState({
        ...appState,
        updateValidationError: true,
        updateError: false,
      });
      return;
    }
    if (!appState.csrfToken) {
      setAppState({
        ...appState,
        updateValidationError: false,
        updateError: true,
      });
      return;
    }

    const csrfTokenToUse = appState.csrfToken;
    setAppState({
      ...appState,
      updatePending: true,
      updateValidationError: false,
      updateError: false,
    });

    const sequence = ++updateSequence.current;
    const updateResult = await client.requestJson<unknown>("/api/v1/settings/instance", {
      method: "PATCH",
      jsonBody: { display_name: normalizedDisplayName },
      csrfToken: csrfTokenToUse,
    });
    if (!isMounted.current || sequence !== updateSequence.current) {
      return;
    }
    if (!updateResult.ok || updateResult.status !== 200 || !parseUpdateInstanceResponse(updateResult.data)) {
      setAppState((previous) => {
        if (previous.kind !== "authorized") {
          return previous;
        }
        return {
          ...previous,
          updatePending: false,
          updateError: true,
        };
      });
      return;
    }

    const settingsResult = await client.requestJson<unknown>("/api/v1/settings");
    if (!isMounted.current || sequence !== updateSequence.current) {
      return;
    }
    if (!settingsResult.ok) {
      if (settingsResult.error.kind === "unauthorized") {
        setAppState({ kind: "unauthenticated" });
        return;
      }
      if (settingsResult.error.kind === "forbidden") {
        setAppState({ kind: "forbidden" });
        return;
      }
      setAppState({ kind: "unavailable" });
      return;
    }

    const refreshedSettings = parsePublicSettingsDto(settingsResult.data);
    if (!refreshedSettings) {
      setAppState({ kind: "unavailable" });
      return;
    }

    setAppState((previous) => {
      if (previous.kind !== "authorized") {
        return previous;
      }
      return {
        ...previous,
        settings: refreshedSettings,
        overview: {
          ...previous.overview,
          instance: {
            ...previous.overview.instance,
            name: refreshedSettings.instance.name,
          },
        },
        editMode: false,
        draftDisplayName: refreshedSettings.instance.name,
        updatePending: false,
        updateError: false,
        updateValidationError: false,
      };
    });
  }, [appState, client]);

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
      if (appState.view === "settings" && appState.settings) {
        return (
          <SettingsShell
            settings={appState.settings}
            editMode={appState.editMode}
            draftDisplayName={appState.draftDisplayName}
            isUpdating={appState.updatePending}
            updateErrorMessage={appState.updateError ? UPDATE_INSTANCE_ERROR_MESSAGE : null}
            updateValidationMessage={
              appState.updateValidationError ? UPDATE_INSTANCE_VALIDATION_MESSAGE : null
            }
            onStartEdit={handleStartEdit}
            onCancelEdit={handleCancelEdit}
            onSaveDisplayName={handleSaveDisplayName}
            onDraftDisplayNameChange={handleDraftDisplayNameChange}
            isSigningOut={appState.logoutPending}
            signOutError={appState.logoutError}
            onSignOut={handleSignOut}
            onBackToOverview={handleBackToOverview}
          />
        );
      }
      return (
        <OverviewShell
          overview={appState.overview}
          isSigningOut={appState.logoutPending}
          signOutError={appState.logoutError}
          settingsAccessError={appState.settingsAccessError}
          settingsLoading={appState.settingsLoading}
          onOpenSettings={handleOpenSettings}
          onSignOut={handleSignOut}
        />
      );
    default:
      return <UnavailableView />;
  }
}

function isPlainRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function hasExactKeys(record: Record<string, unknown>, keys: string[]): boolean {
  const actualKeys = Object.keys(record).sort();
  const expectedKeys = [...keys].sort();
  if (actualKeys.length !== expectedKeys.length) {
    return false;
  }
  return expectedKeys.every((key, index) => key === actualKeys[index]);
}

function parseUpdateInstanceResponse(payload: unknown): { displayName: string } | null {
  if (!isPlainRecord(payload) || !hasExactKeys(payload, ["display_name"])) {
    return null;
  }
  if (typeof payload.display_name !== "string") {
    return null;
  }
  return { displayName: payload.display_name };
}

function validateInstanceDisplayNameDraft(rawValue: string): string | null {
  const trimmed = rawValue.trim();
  if (trimmed.length === 0) {
    return null;
  }
  if (Array.from(trimmed).length > 80) {
    return null;
  }
  if (/[\p{Cc}\p{Cf}]/u.test(trimmed)) {
    return null;
  }
  return trimmed;
}
