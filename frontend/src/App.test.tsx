import "@testing-library/jest-dom/vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import App from "./App";
import type { ApiClient, ApiResult } from "./lib/api";

describe("App", () => {
  function makeApiClientMock() {
    const requestJson = vi.fn();
    const requestNoContent = vi.fn();
    const apiClient: ApiClient = {
      requestJson: requestJson as ApiClient["requestJson"],
      requestNoContent: requestNoContent as NonNullable<ApiClient["requestNoContent"]>,
    };
    return { apiClient, requestJson, requestNoContent };
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

  it("sign out action appears only in authorized settings shell", async () => {
    const { apiClient, requestJson } = makeApiClientMock();
    requestJson
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        requestId: "request-initial-200",
        data: validSettingsPayload(),
      })
      .mockResolvedValueOnce({
        ok: false,
        error: {
          kind: "unauthorized",
          status: 401,
          code: "UNAUTHORIZED",
          message: "Authentication required",
          requestId: "request-401",
        },
      })
      .mockResolvedValueOnce({
        ok: false,
        error: {
          kind: "forbidden",
          status: 403,
          code: "FORBIDDEN",
          message: "Permission denied",
          requestId: "request-403",
        },
      })
      .mockResolvedValueOnce({
        ok: false,
        error: {
          kind: "api_error",
          status: 500,
          code: "INTERNAL_ERROR",
          message: "Internal server error",
          requestId: "request-500",
        },
      });

    const authorized = render(<App apiClient={apiClient} />);
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Sign out" })).toBeInTheDocument();
    });
    authorized.unmount();

    const unauthenticated = render(<App apiClient={apiClient} />);
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
    });
    expect(screen.queryByRole("button", { name: "Sign out" })).not.toBeInTheDocument();
    unauthenticated.unmount();

    const forbidden = render(<App apiClient={apiClient} />);
    await waitFor(() => {
      expect(
        screen.getByText("You do not have access to instance settings.")
      ).toBeInTheDocument();
    });
    expect(screen.queryByRole("button", { name: "Sign out" })).not.toBeInTheDocument();
    forbidden.unmount();

    const unavailable = render(<App apiClient={apiClient} />);
    await waitFor(() => {
      expect(
        screen.getByText("Instance settings are temporarily unavailable.")
      ).toBeInTheDocument();
    });
    expect(screen.queryByRole("button", { name: "Sign out" })).not.toBeInTheDocument();
    unavailable.unmount();
  });

  it("edit name action appears only in authorized settings shell", async () => {
    const { apiClient, requestJson } = makeApiClientMock();
    requestJson
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        requestId: "request-initial-200",
        data: validSettingsPayload(),
      })
      .mockResolvedValueOnce({
        ok: false,
        error: {
          kind: "unauthorized",
          status: 401,
          code: "UNAUTHORIZED",
          message: "Authentication required",
          requestId: "request-401",
        },
      })
      .mockResolvedValueOnce({
        ok: false,
        error: {
          kind: "forbidden",
          status: 403,
          code: "FORBIDDEN",
          message: "Permission denied",
          requestId: "request-403",
        },
      })
      .mockResolvedValueOnce({
        ok: false,
        error: {
          kind: "api_error",
          status: 500,
          code: "INTERNAL_ERROR",
          message: "Internal server error",
          requestId: "request-500",
        },
      });

    const authorized = render(<App apiClient={apiClient} />);
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Edit name" })).toBeInTheDocument();
    });
    authorized.unmount();

    const unauthenticated = render(<App apiClient={apiClient} />);
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
    });
    expect(screen.queryByRole("button", { name: "Edit name" })).not.toBeInTheDocument();
    unauthenticated.unmount();

    const forbidden = render(<App apiClient={apiClient} />);
    await waitFor(() => {
      expect(
        screen.getByText("You do not have access to instance settings.")
      ).toBeInTheDocument();
    });
    expect(screen.queryByRole("button", { name: "Edit name" })).not.toBeInTheDocument();
    forbidden.unmount();

    const unavailable = render(<App apiClient={apiClient} />);
    await waitFor(() => {
      expect(
        screen.getByText("Instance settings are temporarily unavailable.")
      ).toBeInTheDocument();
    });
    expect(screen.queryByRole("button", { name: "Edit name" })).not.toBeInTheDocument();
    unavailable.unmount();
  });

  it("edit mode initializes from server name and cancel discards draft without requests", async () => {
    const { apiClient, requestJson, requestNoContent } = makeApiClientMock();
    requestJson.mockResolvedValue({
      ok: true,
      status: 200,
      requestId: "request-initial-200",
      data: validSettingsPayload(),
    });

    render(<App apiClient={apiClient} />);
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Edit name" })).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: "Edit name" }));
    const input = screen.getByLabelText("Instance name");
    expect(input).toHaveValue("Vyntrio Home");

    fireEvent.change(input, { target: { value: "Draft rename" } });
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));

    expect(screen.queryByLabelText("Instance name")).not.toBeInTheDocument();
    expect(screen.getByText("Vyntrio Home")).toBeInTheDocument();
    expect(screen.queryByText("inert-test-csrf-token")).not.toBeInTheDocument();
    expect(requestJson).toHaveBeenCalledTimes(1);
    expect(requestNoContent).toHaveBeenCalledTimes(0);
  });

  it("save sends exactly one PATCH with json body and csrf, blocks duplicate save, and disables sign out while pending", async () => {
    const { apiClient, requestJson, requestNoContent } = makeApiClientMock();
    let resolvePatch: ((value: ApiResult<unknown>) => void) | undefined;
    const patchPending = new Promise<ApiResult<unknown>>((resolve) => {
      resolvePatch = resolve;
    });

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
      })
      .mockReturnValueOnce(patchPending)
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        requestId: "request-settings-refresh",
        data: {
          ...validSettingsPayload(),
          instance: { ...validSettingsPayload().instance, name: "Renamed Home" },
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
      expect(screen.getByRole("button", { name: "Edit name" })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole("button", { name: "Edit name" }));
    fireEvent.change(screen.getByLabelText("Instance name"), { target: { value: "Renamed Home" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Saving..." })).toBeDisabled();
    });
    expect(screen.getByRole("button", { name: "Sign out" })).toBeDisabled();
    fireEvent.click(screen.getByRole("button", { name: "Saving..." }));
    fireEvent.click(screen.getByRole("button", { name: "Sign out" }));

    expect(requestJson).toHaveBeenCalledTimes(4);
    expect(requestJson.mock.calls[3]).toEqual([
      "/api/v1/settings/instance",
      {
        method: "PATCH",
        jsonBody: { display_name: "Renamed Home" },
        csrfToken: "inert-test-csrf-token",
      },
    ]);
    expect(requestNoContent).toHaveBeenCalledTimes(0);

    resolvePatch?.({
      ok: true,
      status: 200,
      requestId: "request-patch-ok",
      data: { display_name: "Renamed Home" },
    });

    await waitFor(() => {
      expect(screen.getByText("Renamed Home")).toBeInTheDocument();
    });
    expect(requestJson).toHaveBeenCalledTimes(5);
  });

  it("PATCH success triggers exactly one follow-up GET and updates display only from that DTO", async () => {
    const { apiClient, requestJson, requestNoContent } = makeApiClientMock();
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
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        requestId: "request-patch",
        data: { display_name: "Patched Value" },
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
        requestId: "request-refresh",
        data: {
          ...validSettingsPayload(),
          instance: { ...validSettingsPayload().instance, name: "Name From Re-read" },
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
      expect(screen.getByRole("button", { name: "Edit name" })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole("button", { name: "Edit name" }));
    fireEvent.change(screen.getByLabelText("Instance name"), { target: { value: "Patched Value" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(screen.queryByText("Name From Re-read")).not.toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText("Name From Re-read")).toBeInTheDocument();
    });
    expect(screen.queryByText("Patched Value")).not.toBeInTheDocument();

    expect(requestJson).toHaveBeenCalledTimes(5);
    expect(requestJson.mock.calls[4]).toEqual(["/api/v1/settings"]);
    expect(requestJson.mock.calls[4]?.[1]).toBeUndefined();
    expect(requestNoContent).toHaveBeenCalledTimes(0);
  });

  it("PATCH failures keep authorized state, show generic error, clear pending, and allow explicit retry without follow-up GET", async () => {
    const patchFailures: Array<ApiResult<unknown>> = [
      {
        ok: false,
        error: {
          kind: "api_error",
          status: 400,
          code: "BAD_REQUEST",
          message: "Invalid request",
          requestId: "patch-400",
        },
      },
      {
        ok: false,
        error: {
          kind: "unauthorized",
          status: 401,
          code: "UNAUTHORIZED",
          message: "Authentication required",
          requestId: "patch-401",
        },
      },
      {
        ok: false,
        error: {
          kind: "forbidden",
          status: 403,
          code: "FORBIDDEN",
          message: "CSRF validation failed",
          requestId: "patch-403",
        },
      },
      {
        ok: false,
        error: {
          kind: "api_error",
          status: 408,
          code: "REQUEST_TIMEOUT",
          message: "Request timed out",
          requestId: "patch-408",
        },
      },
      {
        ok: false,
        error: {
          kind: "api_error",
          status: 500,
          code: "INTERNAL_ERROR",
          message: "Internal server error",
          requestId: "patch-500",
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
          requestId: "patch-invalid",
        },
      },
      {
        ok: true,
        status: 201,
        requestId: "patch-201",
        data: { display_name: "Renamed Home" },
      },
      {
        ok: true,
        status: 200,
        requestId: "patch-malformed",
        data: { display_name: 42 },
      },
    ];

    for (const patchFailure of patchFailures) {
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
        })
        .mockResolvedValueOnce(patchFailure)
        .mockResolvedValueOnce(patchFailure);

      const view = render(<App apiClient={apiClient} />);
      await waitFor(() => {
        expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
      });
      fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
      fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password-1" } });
      fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

      await waitFor(() => {
        expect(screen.getByRole("button", { name: "Edit name" })).toBeInTheDocument();
      });
      fireEvent.click(screen.getByRole("button", { name: "Edit name" }));
      fireEvent.change(screen.getByLabelText("Instance name"), { target: { value: "Renamed Home" } });
      fireEvent.click(screen.getByRole("button", { name: "Save" }));

      await waitFor(() => {
        expect(
          screen.getByText("The instance name could not be updated. Please try again.")
        ).toBeInTheDocument();
      });
      expect(screen.queryByRole("button", { name: "Saving..." })).not.toBeInTheDocument();
      expect(screen.queryByRole("button", { name: "Sign in" })).not.toBeInTheDocument();

      expect(requestJson).toHaveBeenCalledTimes(4);
      fireEvent.click(screen.getByRole("button", { name: "Save" }));
      expect(requestJson).toHaveBeenCalledTimes(5);
      view.unmount();
    }
  });

  it("missing in-memory csrf blocks PATCH and shows generic update error with no request", async () => {
    const { apiClient, requestJson } = makeApiClientMock();
    requestJson.mockResolvedValue({
      ok: true,
      status: 200,
      requestId: "request-initial-200",
      data: validSettingsPayload(),
    });

    render(<App apiClient={apiClient} />);
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Edit name" })).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: "Edit name" }));
    fireEvent.change(screen.getByLabelText("Instance name"), { target: { value: "Renamed Home" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(requestJson).toHaveBeenCalledTimes(1);
    expect(
      screen.getByText("The instance name could not be updated. Please try again.")
    ).toBeInTheDocument();
  });

  it("follow-up GET failures after PATCH success apply existing session outcome handling", async () => {
    const cases: Array<{
      name: string;
      followup: ApiResult<unknown>;
      expectText: string;
      hideSettings: boolean;
    }> = [
      {
        name: "401 returns to login",
        followup: {
          ok: false,
          error: {
            kind: "unauthorized",
            status: 401,
            code: "UNAUTHORIZED",
            message: "Authentication required",
            requestId: "followup-401",
          },
        },
        expectText: "Sign in",
        hideSettings: true,
      },
      {
        name: "403 returns to forbidden",
        followup: {
          ok: false,
          error: {
            kind: "forbidden",
            status: 403,
            code: "FORBIDDEN",
            message: "Permission denied",
            requestId: "followup-403",
          },
        },
        expectText: "You do not have access to instance settings.",
        hideSettings: true,
      },
      {
        name: "api error returns unavailable",
        followup: {
          ok: false,
          error: {
            kind: "api_error",
            status: 500,
            code: "INTERNAL_ERROR",
            message: "Internal server error",
            requestId: "followup-500",
          },
        },
        expectText: "Instance settings are temporarily unavailable.",
        hideSettings: true,
      },
      {
        name: "network error returns unavailable",
        followup: {
          ok: false,
          error: {
            kind: "network_error",
            status: null,
            code: "NETWORK_ERROR",
            message: "Network request failed",
            requestId: null,
          },
        },
        expectText: "Instance settings are temporarily unavailable.",
        hideSettings: true,
      },
      {
        name: "malformed DTO returns unavailable",
        followup: {
          ok: true,
          status: 200,
          requestId: "followup-bad-dto",
          data: { instance: { name: "bad", version: "x" } },
        },
        expectText: "Instance settings are temporarily unavailable.",
        hideSettings: true,
      },
    ];

    for (const tc of cases) {
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
        })
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          requestId: "request-patch",
          data: { display_name: "Patched Value" },
        })
        .mockResolvedValueOnce(tc.followup);

      const view = render(<App apiClient={apiClient} />);
      await waitFor(() => {
        expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
      });
      fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
      fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password-1" } });
      fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

      await waitFor(() => {
        expect(screen.getByRole("button", { name: "Edit name" })).toBeInTheDocument();
      });
      fireEvent.click(screen.getByRole("button", { name: "Edit name" }));
      fireEvent.change(screen.getByLabelText("Instance name"), { target: { value: "Patched Value" } });
      fireEvent.click(screen.getByRole("button", { name: "Save" }));

      await waitFor(() => {
        if (tc.expectText === "Sign in") {
          expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
          return;
        }
        expect(screen.getByText(tc.expectText)).toBeInTheDocument();
      });
      if (tc.hideSettings) {
        expect(screen.queryByText("Instance settings")).not.toBeInTheDocument();
      }
      expect(screen.queryByText("Patched Value")).not.toBeInTheDocument();
      expect(requestJson).toHaveBeenCalledTimes(5);
      view.unmount();
    }
  });

  it("logout pending disables edit entry and prevents mutation overlap", async () => {
    const { apiClient, requestJson, requestNoContent } = makeApiClientMock();
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

    const logoutPending = new Promise<Awaited<ReturnType<NonNullable<ApiClient["requestNoContent"]>>>>(
      () => {}
    );
    requestNoContent.mockReturnValue(logoutPending);

    render(<App apiClient={apiClient} />);
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
    });
    fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password-1" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Sign out" })).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole("button", { name: "Sign out" }));

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Signing out..." })).toBeDisabled();
    });
    expect(screen.getByRole("button", { name: "Edit name" })).toBeDisabled();
    fireEvent.click(screen.getByRole("button", { name: "Edit name" }));
    expect(screen.queryByLabelText("Instance name")).not.toBeInTheDocument();
    expect(requestJson).toHaveBeenCalledTimes(3);
  });

  it("exact 204 logout sends one POST with csrf option, no body, disables duplicates, and returns to login", async () => {
    const { apiClient, requestJson, requestNoContent } = makeApiClientMock();
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

    let resolveLogout:
      | ((value: Awaited<ReturnType<NonNullable<ApiClient["requestNoContent"]>>>) => void)
      | undefined;
    const logoutPending = new Promise<Awaited<ReturnType<NonNullable<ApiClient["requestNoContent"]>>>>(
      (resolve) => {
        resolveLogout = resolve;
      }
    );
    requestNoContent.mockReturnValue(logoutPending);

    render(<App apiClient={apiClient} />);
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
    });
    fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
    fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password-1" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Sign out" })).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: "Sign out" }));
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Signing out..." })).toBeDisabled();
    });
    fireEvent.click(screen.getByRole("button", { name: "Signing out..." }));

    expect(requestNoContent).toHaveBeenCalledTimes(1);
    expect(requestNoContent).toHaveBeenCalledWith("/api/v1/identity/logout", 204, {
      method: "POST",
      csrfToken: "inert-test-csrf-token",
    });
    expect(requestJson).toHaveBeenCalledTimes(3);

    resolveLogout?.({
      ok: true,
      status: 204,
      requestId: "request-logout",
    });

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
    });
    expect(screen.queryByRole("button", { name: "Sign out" })).not.toBeInTheDocument();
    expect(screen.queryByText("Vyntrio Home")).not.toBeInTheDocument();
    expect(screen.queryByText("inert-test-csrf-token")).not.toBeInTheDocument();
    expect(requestJson).toHaveBeenCalledTimes(3);
    expect(requestNoContent).toHaveBeenCalledTimes(1);
  });

  it("missing in-memory csrf in authorized state does not request logout and shows generic error", async () => {
    const { apiClient, requestJson, requestNoContent } = makeApiClientMock();
    requestJson.mockResolvedValue({
      ok: true,
      status: 200,
      requestId: "request-initial-200",
      data: validSettingsPayload(),
    });

    render(<App apiClient={apiClient} />);
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Sign out" })).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole("button", { name: "Sign out" }));

    expect(requestNoContent).toHaveBeenCalledTimes(0);
    expect(
      screen.getByText("Sign-out could not be completed. Please try again.")
    ).toBeInTheDocument();
    expect(screen.getByText("Vyntrio Home")).toBeInTheDocument();
  });

  it("logout 401 and 403 keep authorized state with generic error and allow manual retry", async () => {
    const failures: Awaited<ReturnType<NonNullable<ApiClient["requestNoContent"]>>>[] = [
      {
        ok: false,
        error: {
          kind: "unauthorized",
          status: 401,
          code: "UNAUTHORIZED",
          message: "Authentication required",
          requestId: "request-logout-401",
        },
      },
      {
        ok: false,
        error: {
          kind: "forbidden",
          status: 403,
          code: "FORBIDDEN",
          message: "CSRF validation failed",
          requestId: "request-logout-403",
        },
      },
    ];

    for (const logoutFailure of failures) {
      const { apiClient, requestJson, requestNoContent } = makeApiClientMock();
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
      requestNoContent.mockResolvedValue(logoutFailure);

      const view = render(<App apiClient={apiClient} />);
      await waitFor(() => {
        expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
      });
      fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
      fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password-1" } });
      fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

      await waitFor(() => {
        expect(screen.getByRole("button", { name: "Sign out" })).toBeInTheDocument();
      });
      fireEvent.click(screen.getByRole("button", { name: "Sign out" }));

      await waitFor(() => {
        expect(
          screen.getByText("Sign-out could not be completed. Please try again.")
        ).toBeInTheDocument();
      });
      expect(screen.getByText("Vyntrio Home")).toBeInTheDocument();
      expect(screen.queryByRole("button", { name: "Sign in" })).not.toBeInTheDocument();
      expect(requestNoContent).toHaveBeenCalledTimes(1);

      fireEvent.click(screen.getByRole("button", { name: "Sign out" }));
      expect(requestNoContent).toHaveBeenCalledTimes(2);
      view.unmount();
    }
  });

  it("logout API/408/500/network/invalid-response/unexpected-2xx failures remain authorized with generic error", async () => {
    const failures: Awaited<ReturnType<NonNullable<ApiClient["requestNoContent"]>>>[] = [
      {
        ok: false,
        error: {
          kind: "api_error",
          status: 408,
          code: "REQUEST_TIMEOUT",
          message: "Request timed out",
          requestId: "request-logout-408",
        },
      },
      {
        ok: false,
        error: {
          kind: "api_error",
          status: 500,
          code: "INTERNAL_ERROR",
          message: "Internal server error",
          requestId: "request-logout-500",
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
          requestId: "request-logout-invalid",
        },
      },
      {
        ok: false,
        error: {
          kind: "invalid_response",
          status: 200,
          code: "INVALID_RESPONSE",
          message: "Invalid server response",
          requestId: "request-logout-200",
        },
      },
      {
        ok: false,
        error: {
          kind: "invalid_response",
          status: 201,
          code: "INVALID_RESPONSE",
          message: "Invalid server response",
          requestId: "request-logout-201",
        },
      },
      {
        ok: false,
        error: {
          kind: "invalid_response",
          status: 202,
          code: "INVALID_RESPONSE",
          message: "Invalid server response",
          requestId: "request-logout-202",
        },
      },
      {
        ok: false,
        error: {
          kind: "invalid_response",
          status: 205,
          code: "INVALID_RESPONSE",
          message: "Invalid server response",
          requestId: "request-logout-205",
        },
      },
    ];

    for (const logoutFailure of failures) {
      const { apiClient, requestJson, requestNoContent } = makeApiClientMock();
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
      requestNoContent.mockResolvedValue(logoutFailure);

      const view = render(<App apiClient={apiClient} />);
      await waitFor(() => {
        expect(screen.getByRole("button", { name: "Sign in" })).toBeInTheDocument();
      });
      fireEvent.change(screen.getByLabelText("Username"), { target: { value: "owner" } });
      fireEvent.change(screen.getByLabelText("Password"), { target: { value: "password-1" } });
      fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

      await waitFor(() => {
        expect(screen.getByRole("button", { name: "Sign out" })).toBeInTheDocument();
      });
      fireEvent.click(screen.getByRole("button", { name: "Sign out" }));

      await waitFor(() => {
        expect(
          screen.getByText("Sign-out could not be completed. Please try again.")
        ).toBeInTheDocument();
      });
      expect(screen.getByText("Vyntrio Home")).toBeInTheDocument();
      expect(screen.queryByText("inert-test-csrf-token")).not.toBeInTheDocument();
      expect(screen.queryByRole("button", { name: "Sign in" })).not.toBeInTheDocument();
      expect(requestNoContent).toHaveBeenCalledTimes(1);
      expect(requestJson).toHaveBeenCalledTimes(3);
      view.unmount();
    }
  });
});
