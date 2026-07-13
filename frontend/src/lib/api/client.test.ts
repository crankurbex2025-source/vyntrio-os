import { describe, expect, it, vi } from "vitest";
import { createApiClient, type ApiResult } from "./client";

function jsonResponse(
  status: number,
  body: unknown,
  headers: Record<string, string> = {}
): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: {
      "Content-Type": "application/json; charset=utf-8",
      ...headers,
    },
  });
}

describe("createApiClient", () => {
  it("uses credentials include and Accept for valid /api/v1 path", async () => {
    const fetchMock = vi.fn<typeof fetch>(async () =>
      jsonResponse(200, { ok: true }, { "X-Request-ID": "req-1" })
    );
    const client = createApiClient(fetchMock);

    await client.requestJson<{ ok: boolean }>("/api/v1/version");

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/version",
      expect.objectContaining({
        method: "GET",
        credentials: "include",
        headers: expect.objectContaining({
          Accept: "application/json",
        }),
      })
    );
  });

  it("sets Content-Type application/json when jsonBody is provided", async () => {
    const fetchMock = vi.fn<typeof fetch>(async () => jsonResponse(200, { ok: true }));
    const client = createApiClient(fetchMock);

    await client.requestJson<{ ok: boolean }>("/api/v1/example", {
      method: "POST",
      jsonBody: { name: "example" },
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/example",
      expect.objectContaining({
        method: "POST",
        headers: expect.objectContaining({
          Accept: "application/json",
          "Content-Type": "application/json",
        }),
        body: JSON.stringify({ name: "example" }),
      })
    );
  });

  it("adds X-CSRF-Token only when explicit non-empty csrfToken is supplied", async () => {
    const fetchMock = vi.fn<typeof fetch>(async () => jsonResponse(200, { ok: true }));
    const client = createApiClient(fetchMock);

    await client.requestJson<{ ok: boolean }>("/api/v1/example", {
      method: "PATCH",
      jsonBody: { display_name: "Home" },
      csrfToken: "placeholder-csrf-token",
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/example",
      expect.objectContaining({
        headers: expect.objectContaining({
          Accept: "application/json",
          "Content-Type": "application/json",
          "X-CSRF-Token": "placeholder-csrf-token",
        }),
      })
    );
  });

  it("does not send X-CSRF-Token header when csrfToken is omitted", async () => {
    const fetchMock = vi.fn<typeof fetch>(async () => jsonResponse(200, { ok: true }));
    const client = createApiClient(fetchMock);

    await client.requestJson<{ ok: boolean }>("/api/v1/example");

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const firstCall = fetchMock.mock.calls[0];
    if (!firstCall) {
      throw new Error("missing fetch call");
    }
    const init = firstCall[1] as RequestInit | undefined;
    const headers = (init?.headers ?? {}) as Record<string, string>;
    expect(headers["X-CSRF-Token"]).toBeUndefined();
  });

  it("rejects unsafe paths before fetch", async () => {
    const fetchMock = vi.fn<typeof fetch>(async () => jsonResponse(200, { ok: true }));
    const client = createApiClient(fetchMock);

    const cases = [
      "https://example.com/api/v1/settings",
      "//example.com/api/v1/settings",
      "/api/v1/settings#fragment",
      "/api/v2/settings",
    ];

    for (const path of cases) {
      const result = await client.requestJson(path);
      expect(result).toEqual({
        ok: false,
        error: {
          kind: "invalid_request",
          status: null,
          code: "INVALID_API_PATH",
          message: "Invalid API path",
          requestId: null,
        },
      });
    }

    expect(fetchMock).toHaveBeenCalledTimes(0);
  });

  it("returns typed JSON success payload and requestId", async () => {
    const fetchMock = vi.fn<typeof fetch>(async () =>
      jsonResponse(
        200,
        { instance: { name: "Vyntrio" } },
        { "X-Request-ID": "request-123" }
      )
    );
    const client = createApiClient(fetchMock);

    const result = await client.requestJson<{ instance: { name: string } }>(
      "/api/v1/settings"
    );

    expect(result).toEqual({
      ok: true,
      status: 200,
      requestId: "request-123",
      data: { instance: { name: "Vyntrio" } },
    });
  });

  it("maps canonical 401 and 403 responses to stable typed errors", async () => {
    const unauthorized = jsonResponse(
      401,
      {
        error: {
          code: "UNAUTHORIZED",
          message: "Authentication required",
          request_id: "req-401",
        },
      },
      { "X-Request-ID": "req-401" }
    );
    const forbidden = jsonResponse(
      403,
      {
        error: {
          code: "FORBIDDEN",
          message: "Permission denied",
          request_id: "req-403",
        },
      },
      { "X-Request-ID": "req-403" }
    );
    const fetchMock = vi.fn<typeof fetch>();
    fetchMock.mockResolvedValueOnce(unauthorized);
    fetchMock.mockResolvedValueOnce(forbidden);
    const client = createApiClient(fetchMock);

    const unauthorizedResult = await client.requestJson("/api/v1/settings");
    const forbiddenResult = await client.requestJson("/api/v1/settings/instance");

    expect(unauthorizedResult).toEqual({
      ok: false,
      error: {
        kind: "unauthorized",
        status: 401,
        code: "UNAUTHORIZED",
        message: "Authentication required",
        requestId: "req-401",
      },
    });
    expect(forbiddenResult).toEqual({
      ok: false,
      error: {
        kind: "forbidden",
        status: 403,
        code: "FORBIDDEN",
        message: "Permission denied",
        requestId: "req-403",
      },
    });
  });

  it("maps canonical non-401/403 API errors safely", async () => {
    const fetchMock = vi.fn<typeof fetch>(async () =>
      jsonResponse(
        400,
        {
          error: {
            code: "BAD_REQUEST",
            message: "Invalid request",
            request_id: "req-400",
          },
        },
        { "X-Request-ID": "req-400" }
      )
    );
    const client = createApiClient(fetchMock);

    const result = await client.requestJson("/api/v1/settings/instance");

    expect(result).toEqual({
      ok: false,
      error: {
        kind: "api_error",
        status: 400,
        code: "BAD_REQUEST",
        message: "Invalid request",
        requestId: "req-400",
      },
    });
  });

  it("maps fetch rejection to stable network error without exposing internal error", async () => {
    const fetchMock = vi.fn<typeof fetch>(async () => {
      throw new Error("connection reset by peer");
    });
    const client = createApiClient(fetchMock);

    const result = await client.requestJson("/api/v1/settings");

    expect(result).toEqual({
      ok: false,
      error: {
        kind: "network_error",
        status: null,
        code: "NETWORK_ERROR",
        message: "Network request failed",
        requestId: null,
      },
    });
  });

  it("maps malformed or non-JSON error responses to invalid_response without raw body retention", async () => {
    const malformedJsonResponse = new Response("{not-json", {
      status: 500,
      headers: {
        "Content-Type": "application/json; charset=utf-8",
        "X-Request-ID": "req-malformed",
      },
    });
    const nonJSONResponse = new Response("<html>error</html>", {
      status: 500,
      headers: {
        "Content-Type": "text/html; charset=utf-8",
        "X-Request-ID": "req-text",
      },
    });

    const fetchMock = vi.fn<typeof fetch>();
    fetchMock.mockResolvedValueOnce(malformedJsonResponse);
    fetchMock.mockResolvedValueOnce(nonJSONResponse);
    const client = createApiClient(fetchMock);

    const malformedResult = await client.requestJson("/api/v1/settings");
    const textResult = await client.requestJson("/api/v1/settings");

    const expectedInvalid = (requestId: string): ApiResult<unknown> => ({
      ok: false,
      error: {
        kind: "invalid_response",
        status: 500,
        code: "INVALID_RESPONSE",
        message: "Invalid server response",
        requestId,
      },
    });

    expect(malformedResult).toEqual(expectedInvalid("req-malformed"));
    expect(textResult).toEqual(expectedInvalid("req-text"));
  });

  it("requestNoContent succeeds only on exact expected status", async () => {
    const fetchMock = vi.fn<typeof fetch>(async () =>
      new Response(null, {
        status: 204,
        headers: { "X-Request-ID": "req-204" },
      })
    );
    const client = createApiClient(fetchMock);

    const result = await client.requestNoContent!("/api/v1/identity/logout", 204, {
      method: "POST",
      csrfToken: "inert-test-csrf-token",
    });

    expect(result).toEqual({
      ok: true,
      status: 204,
      requestId: "req-204",
    });
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/identity/logout",
      expect.objectContaining({
        method: "POST",
        credentials: "include",
        headers: expect.objectContaining({
          Accept: "application/json",
          "X-CSRF-Token": "inert-test-csrf-token",
        }),
      })
    );
    const init = fetchMock.mock.calls[0]?.[1] as RequestInit | undefined;
    expect(init?.body).toBeUndefined();
  });

  it("requestNoContent treats unexpected 2xx as invalid_response", async () => {
    const fetchMock = vi.fn<typeof fetch>(async () =>
      new Response(null, {
        status: 200,
        headers: { "X-Request-ID": "req-200" },
      })
    );
    const client = createApiClient(fetchMock);

    const result = await client.requestNoContent!("/api/v1/identity/logout", 204, {
      method: "POST",
    });

    expect(result).toEqual({
      ok: false,
      error: {
        kind: "invalid_response",
        status: 200,
        code: "INVALID_RESPONSE",
        message: "Invalid server response",
        requestId: "req-200",
      },
    });
  });

  it("requestNoContent maps canonical 401 and 403 errors", async () => {
    const unauthorized = jsonResponse(
      401,
      {
        error: {
          code: "UNAUTHORIZED",
          message: "Authentication required",
          request_id: "req-401",
        },
      },
      { "X-Request-ID": "req-401" }
    );
    const forbidden = jsonResponse(
      403,
      {
        error: {
          code: "FORBIDDEN",
          message: "CSRF validation failed",
          request_id: "req-403",
        },
      },
      { "X-Request-ID": "req-403" }
    );
    const fetchMock = vi.fn<typeof fetch>();
    fetchMock.mockResolvedValueOnce(unauthorized);
    fetchMock.mockResolvedValueOnce(forbidden);
    const client = createApiClient(fetchMock);

    const unauthorizedResult = await client.requestNoContent!("/api/v1/identity/logout", 204, {
      method: "POST",
      csrfToken: "inert-test-csrf-token",
    });
    const forbiddenResult = await client.requestNoContent!("/api/v1/identity/logout", 204, {
      method: "POST",
      csrfToken: "inert-test-csrf-token",
    });

    expect(unauthorizedResult).toEqual({
      ok: false,
      error: {
        kind: "unauthorized",
        status: 401,
        code: "UNAUTHORIZED",
        message: "Authentication required",
        requestId: "req-401",
      },
    });
    expect(forbiddenResult).toEqual({
      ok: false,
      error: {
        kind: "forbidden",
        status: 403,
        code: "FORBIDDEN",
        message: "CSRF validation failed",
        requestId: "req-403",
      },
    });
  });
});
