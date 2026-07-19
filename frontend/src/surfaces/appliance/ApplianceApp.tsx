import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Navigate, Route, Routes, useLocation } from "react-router-dom";
import { LoginScreen } from "../../features/auth/LoginScreen";
import { OverviewShell } from "../../features/overview/OverviewShell";
import { parseOverviewDto, type OverviewDto } from "../../features/overview/overviewDto";
import { SettingsShell } from "../../features/settings/SettingsShell";
import { parsePublicSettingsDto, type PublicSettingsDto } from "../../features/settings/settingsDto";
import { SharesShell } from "../../features/shares/SharesShell";
import { StorageShell } from "../../features/storage/StorageShell";
import { parseStorageLayoutDto, type StorageLayoutDto } from "../../features/storage/storageDto";
import { ToolsShell } from "../../features/tools/ToolsShell";
import { UsersShell } from "../../features/users/UsersShell";
import { createApiClient, type ApiClient } from "../../lib/api";
import { ApplianceShell } from "./ApplianceShell";
import { applianceSectionFromPath } from "./nav/applianceNavConfig";

const UPDATE_INSTANCE_ERROR_MESSAGE = "The instance name could not be updated. Please try again.";
const UPDATE_INSTANCE_VALIDATION_MESSAGE = "Enter a valid instance name.";

type AuthorizedState = {
  kind: "authorized";
  overview: OverviewDto;
  csrfToken: string | null;
  settings: PublicSettingsDto | null;
  storage: StorageLayoutDto | null;
  logoutPending: boolean;
  logoutError: boolean;
  settingsLoading: boolean;
  settingsAccessError: boolean;
  storageLoading: boolean;
  storageAccessError: boolean;
  storageMutationPending: boolean;
  storageMutationError: string | null;
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

type LoginHandoffState = {
  csrfToken?: string;
};

type ApplianceAppProps = {
  apiClient?: ApiClient;
};

function SessionProbeView() {
  return (
    <ApplianceShell>
      <main className="status-wrap">
        <section className="status-card">
          <h1>Vyntrio OS</h1>
          <p>Checking session...</p>
        </section>
      </main>
    </ApplianceShell>
  );
}

function ForbiddenView() {
  return (
    <ApplianceShell>
      <main className="status-wrap">
        <section className="status-card">
          <h1>Vyntrio OS</h1>
          <p>You do not have access to this appliance view.</p>
        </section>
      </main>
    </ApplianceShell>
  );
}

function UnavailableView() {
  return (
    <ApplianceShell>
      <main className="status-wrap">
        <section className="status-card">
          <h1>Vyntrio OS</h1>
          <p>The appliance overview is temporarily unavailable.</p>
        </section>
      </main>
    </ApplianceShell>
  );
}

function SectionLoading({ label }: { label: string }) {
  return (
    <main className="status-wrap">
      <section className="status-card">
        <h1>{label}</h1>
        <p>Loading...</p>
      </section>
    </main>
  );
}

function SectionAccessError({ message }: { message: string }) {
  return (
    <main className="status-wrap">
      <section className="status-card">
        <p role="alert">{message}</p>
      </section>
    </main>
  );
}

function readLoginHandoff(locationState: unknown): string | null {
  if (typeof locationState !== "object" || locationState === null || Array.isArray(locationState)) {
    return null;
  }
  const csrfToken = (locationState as LoginHandoffState).csrfToken;
  if (typeof csrfToken !== "string" || csrfToken.length === 0) {
    return null;
  }
  return csrfToken;
}

export default function ApplianceApp({ apiClient }: ApplianceAppProps) {
  const client = useMemo(() => apiClient ?? createApiClient(), [apiClient]);
  const location = useLocation();
  const section = applianceSectionFromPath(location.pathname);
  const loginHandoffCsrf = useMemo(
    () => readLoginHandoff(location.state),
    [location.state]
  );
  const isMounted = useRef(true);
  const probeSequence = useRef(0);
  const logoutSequence = useRef(0);
  const updateSequence = useRef(0);
  const settingsLoadSequence = useRef(0);
  const storageLoadSequence = useRef(0);
  const [appState, setAppState] = useState<AppState>({ kind: "bootstrapping" });

  const buildAuthorizedState = useCallback(
    (overview: OverviewDto, csrfToken: string | null): AuthorizedState => ({
      kind: "authorized",
      overview,
      csrfToken,
      settings: null,
      storage: null,
      logoutPending: false,
      logoutError: false,
      settingsLoading: false,
      settingsAccessError: false,
      storageLoading: false,
      storageAccessError: false,
      storageMutationPending: false,
      storageMutationError: null,
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
    if (loginHandoffCsrf) {
      setAppState({ kind: "probing_after_login", csrfToken: loginHandoffCsrf });
      void probeOverview("post_login", loginHandoffCsrf);
      return;
    }
    void probeOverview("initial", null);
  }, [loginHandoffCsrf, probeOverview]);

  const handleLoginSuccess = useCallback(
    (csrfToken: string) => {
      setAppState({ kind: "probing_after_login", csrfToken });
      void probeOverview("post_login", csrfToken);
    },
    [probeOverview]
  );

  const loadSettings = useCallback(async () => {
    if (appState.kind !== "authorized") {
      return;
    }
    if (
      appState.settings ||
      appState.settingsLoading ||
      appState.settingsAccessError ||
      appState.logoutPending
    ) {
      return;
    }

    const sequence = ++settingsLoadSequence.current;
    setAppState((previous) => {
      if (previous.kind !== "authorized") {
        return previous;
      }
      return {
        ...previous,
        settingsLoading: true,
        settingsAccessError: false,
      };
    });

    const result = await client.requestJson<unknown>("/api/v1/settings");
    if (!isMounted.current || sequence !== settingsLoadSequence.current) {
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
            settings: null,
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

  const loadStorage = useCallback(async () => {
    if (appState.kind !== "authorized") {
      return;
    }
    if (
      appState.storage ||
      appState.storageLoading ||
      appState.storageAccessError ||
      appState.logoutPending
    ) {
      return;
    }

    const sequence = ++storageLoadSequence.current;
    setAppState((previous) => {
      if (previous.kind !== "authorized") {
        return previous;
      }
      return {
        ...previous,
        storageLoading: true,
        storageAccessError: false,
      };
    });

    const [disksResult, poolsResult, sharesResult] = await Promise.all([
      client.requestJson<unknown>("/api/v1/storage/disks"),
      client.requestJson<unknown>("/api/v1/storage/pools"),
      client.requestJson<unknown>("/api/v1/storage/shares"),
    ]);
    if (!isMounted.current || sequence !== storageLoadSequence.current) {
      return;
    }

    const firstError = [disksResult, poolsResult, sharesResult].find((result) => !result.ok);
    if (firstError && !firstError.ok) {
      if (firstError.error.kind === "unauthorized") {
        setAppState({ kind: "unauthenticated" });
        return;
      }
      if (firstError.error.kind === "forbidden") {
        setAppState((previous) => {
          if (previous.kind !== "authorized") {
            return previous;
          }
          return {
            ...previous,
            storageLoading: false,
            storageAccessError: true,
            storage: null,
          };
        });
        return;
      }
      setAppState({ kind: "unavailable" });
      return;
    }

    const storage = parseStorageLayoutDto({
      disks: disksResult.ok ? disksResult.data : null,
      pools: poolsResult.ok ? poolsResult.data : null,
      shares: sharesResult.ok ? sharesResult.data : null,
    });
    if (!storage) {
      setAppState({ kind: "unavailable" });
      return;
    }

    setAppState((previous) => {
      if (previous.kind !== "authorized") {
        return previous;
      }
      return {
        ...previous,
        storage,
        storageLoading: false,
        storageAccessError: false,
      };
    });
  }, [appState, client]);

  useEffect(() => {
    if (appState.kind !== "authorized") {
      return;
    }
    if (section === "settings") {
      void loadSettings();
    }
    if (section === "storage" || section === "shares") {
      void loadStorage();
    }
  }, [appState.kind, section, loadSettings, loadStorage]);

  const refreshStorageLayout = useCallback(async (): Promise<StorageLayoutDto | null> => {
    const [disksResult, poolsResult, sharesResult] = await Promise.all([
      client.requestJson<unknown>("/api/v1/storage/disks"),
      client.requestJson<unknown>("/api/v1/storage/pools"),
      client.requestJson<unknown>("/api/v1/storage/shares"),
    ]);
    if (![disksResult, poolsResult, sharesResult].every((result) => result.ok)) {
      return null;
    }
    return parseStorageLayoutDto({
      disks: disksResult.ok ? disksResult.data : null,
      pools: poolsResult.ok ? poolsResult.data : null,
      shares: sharesResult.ok ? sharesResult.data : null,
    });
  }, [client]);

  const handleCreatePool = useCallback(
    async (name: string, diskIds: string[]) => {
      if (appState.kind !== "authorized" || !appState.csrfToken || !appState.storage) {
        return;
      }
      setAppState({ ...appState, storageMutationPending: true, storageMutationError: null });
      const result = await client.requestJson<unknown>("/api/v1/storage/pools", {
        method: "POST",
        jsonBody: { name, disk_ids: diskIds, confirm: true },
        csrfToken: appState.csrfToken,
      });
      if (!isMounted.current) {
        return;
      }
      if (!result.ok) {
        setAppState((previous) => {
          if (previous.kind !== "authorized") {
            return previous;
          }
          return {
            ...previous,
            storageMutationPending: false,
            storageMutationError:
              result.error.kind === "api_error"
                ? result.error.message
                : "Pool could not be declared.",
          };
        });
        return;
      }
      const layout = await refreshStorageLayout();
      setAppState((previous) => {
        if (previous.kind !== "authorized") {
          return previous;
        }
        return {
          ...previous,
          storage: layout ?? previous.storage,
          storageMutationPending: false,
          storageMutationError: layout ? null : "Pool declared but layout refresh failed.",
        };
      });
    },
    [appState, client, refreshStorageLayout]
  );

  const handleAddDataset = useCallback(
    async (poolId: string, name: string) => {
      if (appState.kind !== "authorized" || !appState.csrfToken) {
        return;
      }
      setAppState({ ...appState, storageMutationPending: true, storageMutationError: null });
      const result = await client.requestJson<unknown>(`/api/v1/storage/pools/${poolId}/datasets`, {
        method: "POST",
        jsonBody: { name },
        csrfToken: appState.csrfToken,
      });
      if (!isMounted.current) {
        return;
      }
      if (!result.ok) {
        setAppState((previous) => {
          if (previous.kind !== "authorized") {
            return previous;
          }
          return {
            ...previous,
            storageMutationPending: false,
            storageMutationError:
              result.error.kind === "api_error"
                ? result.error.message
                : "Dataset could not be prepared.",
          };
        });
        return;
      }
      const layout = await refreshStorageLayout();
      setAppState((previous) => {
        if (previous.kind !== "authorized") {
          return previous;
        }
        return {
          ...previous,
          storage: layout ?? previous.storage,
          storageMutationPending: false,
          storageMutationError: layout ? null : "Dataset prepared but layout refresh failed.",
        };
      });
    },
    [appState, client, refreshStorageLayout]
  );

  const handleCreateShare = useCallback(
    async (name: string, poolId: string, datasetId?: string) => {
      if (appState.kind !== "authorized" || !appState.csrfToken) {
        return;
      }
      setAppState({ ...appState, storageMutationPending: true, storageMutationError: null });
      const result = await client.requestJson<unknown>("/api/v1/storage/shares", {
        method: "POST",
        jsonBody: { name, pool_id: poolId, dataset_id: datasetId, protocol: "planned" },
        csrfToken: appState.csrfToken,
      });
      if (!isMounted.current) {
        return;
      }
      if (!result.ok) {
        setAppState((previous) => {
          if (previous.kind !== "authorized") {
            return previous;
          }
          return {
            ...previous,
            storageMutationPending: false,
            storageMutationError:
              result.error.kind === "api_error"
                ? result.error.message
                : "Share plan could not be prepared.",
          };
        });
        return;
      }
      const layout = await refreshStorageLayout();
      setAppState((previous) => {
        if (previous.kind !== "authorized") {
          return previous;
        }
        return {
          ...previous,
          storage: layout ?? previous.storage,
          storageMutationPending: false,
          storageMutationError: layout ? null : "Share prepared but layout refresh failed.",
        };
      });
    },
    [appState, client, refreshStorageLayout]
  );

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
      if (previous.kind !== "authorized" || !previous.settings) {
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
    if (appState.kind !== "authorized" || !appState.settings) {
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
      return (
        <ApplianceShell>
          <LoginScreen apiClient={client} onLoginSuccess={handleLoginSuccess} />
        </ApplianceShell>
      );
    case "forbidden":
      return <ForbiddenView />;
    case "unavailable":
      return <UnavailableView />;
    case "authorized":
      return (
        <ApplianceShell
          withNav
          instanceName={appState.overview.instance.name}
          isSigningOut={appState.logoutPending}
          signOutDisabled={appState.updatePending || appState.storageMutationPending}
          onSignOut={handleSignOut}
        >
          {appState.logoutError ? (
            <section className="dashboard-alert" role="alert">
              Sign-out could not be completed. Please try again.
            </section>
          ) : null}
          <Routes>
            <Route
              index
              element={
                <OverviewShell
                  overview={appState.overview}
                  signOutError={false}
                  settingsAccessError={appState.settingsAccessError}
                  storageAccessError={appState.storageAccessError}
                  settingsLoading={appState.settingsLoading}
                  storageLoading={appState.storageLoading}
                />
              }
            />
            <Route
              path="storage"
              element={
                appState.storageAccessError ? (
                  <SectionAccessError message="You do not have access to the storage inventory." />
                ) : appState.storage ? (
                  <StorageShell
                    layout={appState.storage}
                    mutationPending={appState.storageMutationPending}
                    mutationError={appState.storageMutationError}
                    onCreatePool={handleCreatePool}
                    onAddDataset={handleAddDataset}
                  />
                ) : (
                  <SectionLoading label="Storage" />
                )
              }
            />
            <Route
              path="shares"
              element={
                appState.storageAccessError ? (
                  <SectionAccessError message="You do not have access to share plans." />
                ) : appState.storage ? (
                  <SharesShell
                    layout={appState.storage}
                    mutationPending={appState.storageMutationPending}
                    mutationError={appState.storageMutationError}
                    onCreateShare={handleCreateShare}
                  />
                ) : (
                  <SectionLoading label="Shares" />
                )
              }
            />
            <Route path="users" element={<UsersShell />} />
            <Route
              path="settings"
              element={
                appState.settingsAccessError ? (
                  <SectionAccessError message="You do not have access to instance settings." />
                ) : appState.settings ? (
                  <SettingsShell
                    settings={appState.settings}
                    editMode={appState.editMode}
                    draftDisplayName={appState.draftDisplayName}
                    isUpdating={appState.updatePending}
                    updateErrorMessage={
                      appState.updateError ? UPDATE_INSTANCE_ERROR_MESSAGE : null
                    }
                    updateValidationMessage={
                      appState.updateValidationError
                        ? UPDATE_INSTANCE_VALIDATION_MESSAGE
                        : null
                    }
                    onStartEdit={handleStartEdit}
                    onCancelEdit={handleCancelEdit}
                    onSaveDisplayName={handleSaveDisplayName}
                    onDraftDisplayNameChange={handleDraftDisplayNameChange}
                    controlsLocked={appState.logoutPending}
                  />
                ) : (
                  <SectionLoading label="Settings" />
                )
              }
            />
            <Route path="tools" element={<ToolsShell />} />
            <Route path="*" element={<Navigate to="/app" replace />} />
          </Routes>
        </ApplianceShell>
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
