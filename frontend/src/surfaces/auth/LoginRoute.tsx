import { useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { LoginScreen } from "../../features/auth/LoginScreen";
import { createApiClient, type ApiClient } from "../../lib/api";
import { ApplianceShell } from "../appliance/ApplianceShell";

type LoginRouteProps = {
  apiClient?: ApiClient;
};

export function LoginRoute({ apiClient }: LoginRouteProps) {
  const navigate = useNavigate();
  const client = useMemo(() => apiClient ?? createApiClient(), [apiClient]);

  return (
    <ApplianceShell>
      <LoginScreen
        apiClient={client}
        onLoginSuccess={(csrfToken) => {
          navigate("/app", { replace: true, state: { csrfToken } });
        }}
      />
    </ApplianceShell>
  );
}
