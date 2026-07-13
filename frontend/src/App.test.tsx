import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import App from "./App";
import type { ApiClient, ApiResult } from "./lib/api";

describe("App", () => {
  function makeApiClientMock() {
    const requestJson = vi.fn();
    const apiClient: ApiClient = {
      requestJson: requestJson as ApiClient["requestJson"],
    };
    return { apiClient, requestJson };
  }

  function validSettingsPayload() {
    return {
      instance: {
        name: "Vyntrio Home",
        version: "0.2.0-dev",
      },
      api: {
        environment: "development",
      },
    };
  }

  it("initial mount shows checking state and issues exactly one GET /api/v1/settings with no body or csrf", async () => {
    const { apiClient, requestJson } = makeApiClientMock();
    const pending = new Promise<ApiResult<unknown>>(() => {});
    requestJson.mockReturnValue(pending);

    render(<App apiClient={apiClient} />);

    expect(screen.getByText("Checking session...")).toBeInTheDocument();
    expect(screen.queryByLabelText("Username")).not.toBeInTheDocument();
    expect(screen.queryByText("Instance settings")).not.toBeInTheDocument();

    await waitFor(() => {
      expect(requestJson).toHaveBeenCalledTimes(1);
    });
    expect(requestJson).toHaveBeenCalledWith("/api/v1/settings");
    expect(requestJson.mock.calls[0]?.[1]).toBeUndefined();
  });

  it("initial 200 with valid DTO enters authorized readonly settings shell", async () => {
    const { apiClient, requestJson } = makeApiClientMock();
    requestJson.mockResolvedValue({
      ok: true,
      status: 200,
      requestId: "request-ok",
      data: validSettingsPayload(),
    });

    render(<App apiClient={apiClient} />);

    await waitFor(() => {
      expect(screen.getByRole("heading", { name: "Instance settings" })).toBeInTheDocument();
    });
    expect(screen.getByText("Vyntrio Home")).toBeInTheDocument();
    expect(screen.getByText("0.2.0-dev")).toBeInTheDocument();
    expect(screen.getByText("development")).toBeInTheDocument();
    expect(screen.queryByLabelText("Username")).not.toBeInTheDocument();
    expect(screen.queryByText("csrf_token")).not.toBeInTheDocument();
    expect(screen.queryByText("session")).not.toBeInTheDocument();
    expect(screen.queryByText("role")).not.toBeInTheDocument();
  });

  it("initial 401 enters LoginScreen and does not auto-retry", async () => {
    const { apiClient, requestJson } = makeApiClientMock();
    requestJson.mockResolvedValue({
      ok: false,
      error: {
        kind: "unauthorized",
        status: 401,
        code: "UNAUTHORIZED",
        message: "Authentication required",
        requestId: "request-401",
      },
    });

    render(<App apiClient={apiClient} />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
    });
    expect(screen.queryByText("Instance settings")).not.toBeInTheDocument();
    expect(requestJson).toHaveBeenCalledTimes(1);
  });

  it("initial 403 enters forbidden state with no protected data and no retry", async () => {
    const { apiClient, requestJson } = makeApiClientMock();
    requestJson.mockResolvedValue({
      ok: false,
      error: {
        kind: "forbidden",
        status: 403,
        code: "FORBIDDEN",
        message: "Permission denied",
        requestId: "request-403",
      },
    });

    render(<App apiClient={apiClient} />);

    await waitFor(() => {
      expect(
        screen.getByText("You do not have access to instance settings.")
      ).toBeInTheDocument();
    });
    expect(screen.queryByText("Instance settings")).not.toBeInTheDocument();
    expect(screen.queryByText("owner")).not.toBeInTheDocument();
    expect(requestJson).toHaveBeenCalledTimes(1);
  });

  it("initial API/network/invalid-response/malformed-200 enter same unavailable state without retry", async () => {
    const failures: Array<ApiResult<unknown>> = [
      {
        ok: false,
        error: {
          kind: "api_error",
          status: 500,
          code: "INTERNAL_ERROR",
          message: "Internal server error",
          requestId: "request-500",
        },
      },
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
      {
        ok: true,
        status: 200,
        requestId: "request-malformed",
        data: {
          instance: { name: "bad", version: "x" },
        },
      },
    ];

    for (const failure of failures) {
      const { apiClient, requestJson } = makeApiClientMock();
      requestJson.mockResolvedValue(failure);
      const view = render(<App apiClient={apiClient} />);

      await waitFor(() => {
        expect(
          screen.getByText("Instance settings are temporarily unavailable.")
        ).toBeInTheDocument();
      });
      expect(screen.queryByText("Instance settings")).not.toBeInTheDocument();
      expect(screen.queryByText("Vyntrio Home")).not.toBeInTheDocument();
      expect(requestJson).toHaveBeenCalledTimes(1);
      view.unmount();
    }
  });

  it("successful login triggers exactly one post-login GET and enters authorized only after valid settings 200", async () => {
    const { apiClient, requestJson } = makeApiClientMock();
    requestJson
      .mockResolvedValueOnce({
        ok: false,
        error: {
          kind: "unauthorized",
          status: 401,
          code: "UNAUTHORIZED",
          message: "Authentication required",
          requestId: "request-initial-401",
        },
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        requestId: "request-login",
        data: { csrf_token: "inert-test-csrf-token" },
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        requestId: "request-settings",
        data: validSettingsPayload(),
      });

    render(<App apiClient={apiClient} />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password-1" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await waitFor(() => {
      expect(screen.getByText("Checking session...")).toBeInTheDocument();
    });

    await waitFor(() => {
      expect(screen.getByRole("heading", { name: "Instance settings" })).toBeInTheDocument();
    });
    expect(screen.queryByText("inert-test-csrf-token")).not.toBeInTheDocument();

    expect(requestJson).toHaveBeenCalledTimes(3);
    expect(requestJson.mock.calls[0]).toEqual(["/api/v1/settings"]);
    expect(requestJson.mock.calls[1]).toEqual([
      "/api/v1/identity/login",
      {
        method: "POST",
        jsonBody: {
          username: "owner",
          password: "password-1",
        },
      },
    ]);
    expect(requestJson.mock.calls[2]).toEqual(["/api/v1/settings"]);

    const methods = requestJson.mock.calls.map((call) => call[1]?.method);
    expect(methods.filter((method) => method === "POST")).toHaveLength(1);
    expect(methods.filter((method) => method && method !== "POST")).toHaveLength(0);
  });

  it("post-login 401 returns to LoginScreen with no protected data", async () => {
    const { apiClient, requestJson } = makeApiClientMock();
    requestJson
      .mockResolvedValueOnce({
        ok: false,
        error: {
          kind: "unauthorized",
          status: 401,
          code: "UNAUTHORIZED",
          message: "Authentication required",
          requestId: "request-initial-401",
        },
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        requestId: "request-login",
        data: { csrf_token: "inert-test-csrf-token" },
      })
      .mockResolvedValueOnce({
        ok: false,
        error: {
          kind: "unauthorized",
          status: 401,
          code: "UNAUTHORIZED",
          message: "Authentication required",
          requestId: "request-post-login-401",
        },
      });

    render(<App apiClient={apiClient} />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password-1" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
    });
    expect(screen.queryByText("Instance settings")).not.toBeInTheDocument();
    expect(screen.queryByText("inert-test-csrf-token")).not.toBeInTheDocument();
  });

  it("post-login 403 enters forbidden state with no settings, role, or token", async () => {
    const { apiClient, requestJson } = makeApiClientMock();
    requestJson
      .mockResolvedValueOnce({
        ok: false,
        error: {
          kind: "unauthorized",
          status: 401,
          code: "UNAUTHORIZED",
          message: "Authentication required",
          requestId: "request-initial-401",
        },
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        requestId: "request-login",
        data: { csrf_token: "inert-test-csrf-token" },
      })
      .mockResolvedValueOnce({
        ok: false,
        error: {
          kind: "forbidden",
          status: 403,
          code: "FORBIDDEN",
          message: "Permission denied",
          requestId: "request-post-login-403",
        },
      });

    render(<App apiClient={apiClient} />);

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
    });

    fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password-1" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await waitFor(() => {
      expect(
        screen.getByText("You do not have access to instance settings.")
      ).toBeInTheDocument();
    });
    expect(screen.queryByText("Instance settings")).not.toBeInTheDocument();
    expect(screen.queryByText("owner")).not.toBeInTheDocument();
    expect(screen.queryByText("inert-test-csrf-token")).not.toBeInTheDocument();
  });

  it("post-login API/network/invalid-response/malformed-DTO failures enter unavailable state", async () => {
    const postLoginFailures: Array<ApiResult<unknown>> = [
      {
        ok: false,
        error: {
          kind: "api_error",
          status: 500,
          code: "INTERNAL_ERROR",
          message: "Internal server error",
          requestId: "request-post-login-500",
        },
      },
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
          requestId: "request-post-login-invalid",
        },
      },
      {
        ok: true,
        status: 200,
        requestId: "request-post-login-malformed",
        data: {
          instance: { name: "bad", version: "x" },
          api: { environment: "dev", extra: "x" },
        },
      },
    ];

    for (const failure of postLoginFailures) {
      const { apiClient, requestJson } = makeApiClientMock();
      requestJson
        .mockResolvedValueOnce({
          ok: false,
          error: {
            kind: "unauthorized",
            status: 401,
            code: "UNAUTHORIZED",
            message: "Authentication required",
            requestId: "request-initial-401",
          },
        })
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          requestId: "request-login",
          data: { csrf_token: "inert-test-csrf-token" },
        })
        .mockResolvedValueOnce(failure);

      const view = render(<App apiClient={apiClient} />);

      await waitFor(() => {
        expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
      });

      fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
      fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password-1" } });
      fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

      await waitFor(() => {
        expect(
          screen.getByText("Instance settings are temporarily unavailable.")
        ).toBeInTheDocument();
      });
      expect(screen.queryByText("Instance settings")).not.toBeInTheDocument();
      expect(screen.queryByText("inert-test-csrf-token")).not.toBeInTheDocument();
      view.unmount();
    }
  });
});
