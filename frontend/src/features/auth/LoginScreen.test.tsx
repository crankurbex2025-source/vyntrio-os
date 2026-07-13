import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { ApiClient, ApiResult } from "../../lib/api";
import { LoginScreen } from "./LoginScreen";

function makeApiClientMock() {
  const requestJson = vi.fn();
  const apiClient: ApiClient = {
    requestJson: requestJson as ApiClient["requestJson"],
  };
  return {
    apiClient,
    requestJson,
  };
}

function renderLoginScreen(apiClient: ApiClient) {
  render(<LoginScreen apiClient={apiClient} />);
}

describe("LoginScreen", () => {
  it("renders login controls and makes no request on initial render", () => {
    const { apiClient, requestJson } = makeApiClientMock();
    renderLoginScreen(apiClient);

    expect(screen.getByLabelText("Username")).toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
    expect(requestJson).not.toHaveBeenCalled();
  });

  it("submits exactly one login POST request shape through api client with no csrf option", async () => {
    const { apiClient, requestJson } = makeApiClientMock();
    const loginSuccess: ApiResult<unknown> = {
      ok: true,
      status: 200,
      requestId: "request-123",
      data: { csrf_token: "inert-test-csrf-token" },
    };
    requestJson.mockResolvedValue(loginSuccess);
    renderLoginScreen(apiClient);

    fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "safe-password" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await waitFor(() => {
      expect(requestJson).toHaveBeenCalledTimes(1);
    });
    expect(requestJson).toHaveBeenCalledWith("/api/v1/identity/login", {
      method: "POST",
      jsonBody: {
        username: "owner",
        password: "safe-password",
      },
    });
  });

  it("prevents duplicate submit while request is pending", async () => {
    const { apiClient, requestJson } = makeApiClientMock();

    let resolveRequest: ((value: ApiResult<unknown>) => void) | undefined;
    const pendingPromise = new Promise<ApiResult<unknown>>((resolve) => {
      resolveRequest = resolve;
    });
    requestJson.mockReturnValue(pendingPromise);
    renderLoginScreen(apiClient);

    fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "pending-pass" } });

    const submitButton = screen.getByRole("button", { name: "Sign in" });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Signing in..." })).toBeDisabled();
    });
    fireEvent.click(screen.getByRole("button", { name: "Signing in..." }));
    expect(requestJson).toHaveBeenCalledTimes(1);

    resolveRequest?.({
      ok: true,
      status: 200,
      requestId: "request-pending",
      data: { csrf_token: "inert-test-csrf-token" },
    });

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
    });
  });

  it("accepts exact success payload and calls onLoginSuccess without rendering csrf token", async () => {
    const { apiClient, requestJson } = makeApiClientMock();
    const onLoginSuccess = vi.fn();
    requestJson.mockResolvedValue({
      ok: true,
      status: 200,
      requestId: "request-ok",
      data: { csrf_token: "inert-test-csrf-token" },
    });
    render(<LoginScreen apiClient={apiClient} onLoginSuccess={onLoginSuccess} />);

    fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "secret-pass" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await waitFor(() => {
      expect(onLoginSuccess).toHaveBeenCalledWith("inert-test-csrf-token");
    });
    expect(screen.queryByText("inert-test-csrf-token")).not.toBeInTheDocument();
    expect(requestJson).toHaveBeenCalledTimes(1);
    expect(requestJson).toHaveBeenCalledWith(
      "/api/v1/identity/login",
      expect.any(Object)
    );
  });

  it("shows only generic error for backend authentication failure and clears password", async () => {
    const { apiClient, requestJson } = makeApiClientMock();
    requestJson.mockResolvedValue({
      ok: false,
      error: {
        kind: "unauthorized",
        status: 401,
        code: "UNAUTHORIZED",
        message: "Authentication failed",
        requestId: "request-auth-fail",
      },
    });
    renderLoginScreen(apiClient);

    fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "wrong-pass" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await waitFor(() => {
      expect(
        screen.getByText("Sign-in failed. Check your credentials and try again.")
      ).toBeInTheDocument();
    });

    expect(screen.queryByText("Authentication failed")).not.toBeInTheDocument();
    expect(screen.queryByText("UNAUTHORIZED")).not.toBeInTheDocument();
    expect(screen.queryByText("request-auth-fail")).not.toBeInTheDocument();
    expect(screen.queryByText("401")).not.toBeInTheDocument();
    expect((screen.getByLabelText("Password") as HTMLInputElement).value).toBe("");
  });

  it("shows same generic error for network and invalid-response failures and clears password", async () => {
    const cases: ApiResult<unknown>[] = [
      {
        ok: false,
        error: {
          kind: "network_error",
          status: null,
          code: "NETWORK_ERROR",
          message: "Network request failed",
          requestId: null,
        },
      },
      {
        ok: false,
        error: {
          kind: "invalid_response",
          status: 500,
          code: "INVALID_RESPONSE",
          message: "Invalid server response",
          requestId: "request-invalid",
        },
      },
    ];

    for (const failureResult of cases) {
      const { apiClient, requestJson } = makeApiClientMock();
      requestJson.mockResolvedValue(failureResult);
      const view = render(<LoginScreen apiClient={apiClient} />);

      fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
      fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password-1" } });
      fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

      await waitFor(() => {
        expect(
          screen.getByText("Sign-in failed. Check your credentials and try again.")
        ).toBeInTheDocument();
      });
      expect((screen.getByLabelText("Password") as HTMLInputElement).value).toBe("");
      view.unmount();
    }
  });

  it("rejects malformed success payloads safely", async () => {
    const malformedPayloads: unknown[] = [
      {},
      { csrf_token: "" },
      { csrf_token: 123 },
      { csrf_token: "inert-test-csrf-token", unexpected: "value" },
    ];

    for (const data of malformedPayloads) {
      const { apiClient, requestJson } = makeApiClientMock();
      requestJson.mockResolvedValue({
        ok: true,
        status: 200,
        requestId: "request-malformed",
        data,
      });
      const view = render(<LoginScreen apiClient={apiClient} />);

      fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
      fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password-1" } });
      fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

      await waitFor(() => {
        expect(
          screen.getByText("Sign-in failed. Check your credentials and try again.")
        ).toBeInTheDocument();
      });
      expect((screen.getByLabelText("Password") as HTMLInputElement).value).toBe("");
      view.unmount();
    }
  });

  it("does not auto-fetch settings or other routes on initial render", () => {
    const { apiClient, requestJson } = makeApiClientMock();
    renderLoginScreen(apiClient);

    expect(screen.getByLabelText("Username")).toBeInTheDocument();
    expect(screen.getByLabelText("Password")).toBeInTheDocument();
    expect(requestJson).toHaveBeenCalledTimes(0);
  });
});
