export type ApiMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

export type ApiRequestOptions = {
  method?: ApiMethod;
  jsonBody?: unknown;
  csrfToken?: string;
};

export type ApiSuccess<T> = {
  ok: true;
  status: number;
  requestId: string | null;
  data: T;
};

export type ApiErrorKind =
  | "invalid_request"
  | "unauthorized"
  | "forbidden"
  | "api_error"
  | "network_error"
  | "invalid_response";

export type ApiError = {
  kind: ApiErrorKind;
  status: number | null;
  code: string;
  message: string;
  requestId: string | null;
};

export type ApiFailure = {
  ok: false;
  error: ApiError;
};

export type ApiResult<T> = ApiSuccess<T> | ApiFailure;

type FetchLike = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>;

type CanonicalApiErrorEnvelope = {
  error: {
    code: string;
    message: string;
    request_id: string;
  };
};

const API_BASE_PATH = "/api/v1";
const REQUEST_ID_HEADER = "X-Request-ID";
const ACCEPT_HEADER_VALUE = "application/json";
const CONTENT_TYPE_JSON = "application/json";

function makeError(error: ApiError): ApiFailure {
  return { ok: false, error };
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isCanonicalApiErrorEnvelope(value: unknown): value is CanonicalApiErrorEnvelope {
  if (!isObject(value) || !isObject(value.error)) {
    return false;
  }

  return (
    typeof value.error.code === "string" &&
    typeof value.error.message === "string" &&
    typeof value.error.request_id === "string"
  );
}

function isJsonContentType(contentType: string | null): boolean {
  if (!contentType) {
    return false;
  }
  return contentType.toLowerCase().includes(CONTENT_TYPE_JSON);
}

function validateApiPath(path: string): ApiError | null {
  if (typeof path !== "string" || path.trim() === "") {
    return {
      kind: "invalid_request",
      status: null,
      code: "INVALID_API_PATH",
      message: "Invalid API path",
      requestId: null,
    };
  }

  if (path.includes("://") || path.startsWith("//") || path.includes("#") || path.includes("?")) {
    return {
      kind: "invalid_request",
      status: null,
      code: "INVALID_API_PATH",
      message: "Invalid API path",
      requestId: null,
    };
  }

  if (!path.startsWith("/")) {
    return {
      kind: "invalid_request",
      status: null,
      code: "INVALID_API_PATH",
      message: "Invalid API path",
      requestId: null,
    };
  }

  if (!(path === API_BASE_PATH || path.startsWith(`${API_BASE_PATH}/`))) {
    return {
      kind: "invalid_request",
      status: null,
      code: "INVALID_API_PATH",
      message: "Invalid API path",
      requestId: null,
    };
  }

  return null;
}

export type ApiClient = {
  requestJson<T>(path: string, options?: ApiRequestOptions): Promise<ApiResult<T>>;
};

export function createApiClient(fetchImpl: FetchLike = fetch): ApiClient {
  async function requestJson<T>(
    path: string,
    options: ApiRequestOptions = {}
  ): Promise<ApiResult<T>> {
    const pathError = validateApiPath(path);
    if (pathError) {
      return makeError(pathError);
    }

    const method = options.method ?? "GET";
    const headers: Record<string, string> = {
      Accept: ACCEPT_HEADER_VALUE,
    };

    if (options.jsonBody !== undefined) {
      headers["Content-Type"] = CONTENT_TYPE_JSON;
    }

    if (typeof options.csrfToken === "string" && options.csrfToken.trim() !== "") {
      headers["X-CSRF-Token"] = options.csrfToken;
    }

    let response: Response;
    try {
      response = await fetchImpl(path, {
        method,
        credentials: "include",
        headers,
        body: options.jsonBody !== undefined ? JSON.stringify(options.jsonBody) : undefined,
      });
    } catch {
      return makeError({
        kind: "network_error",
        status: null,
        code: "NETWORK_ERROR",
        message: "Network request failed",
        requestId: null,
      });
    }

    const responseRequestID = response.headers.get(REQUEST_ID_HEADER);
    const contentType = response.headers.get("Content-Type");

    if (response.ok) {
      if (!isJsonContentType(contentType)) {
        return makeError({
          kind: "invalid_response",
          status: response.status,
          code: "INVALID_RESPONSE",
          message: "Invalid server response",
          requestId: responseRequestID,
        });
      }

      let payload: unknown;
      try {
        payload = await response.json();
      } catch {
        return makeError({
          kind: "invalid_response",
          status: response.status,
          code: "INVALID_RESPONSE",
          message: "Invalid server response",
          requestId: responseRequestID,
        });
      }

      return {
        ok: true,
        status: response.status,
        requestId: responseRequestID,
        data: payload as T,
      };
    }

    if (!isJsonContentType(contentType)) {
      return makeError({
        kind: "invalid_response",
        status: response.status,
        code: "INVALID_RESPONSE",
        message: "Invalid server response",
        requestId: responseRequestID,
      });
    }

    let errorPayload: unknown;
    try {
      errorPayload = await response.json();
    } catch {
      return makeError({
        kind: "invalid_response",
        status: response.status,
        code: "INVALID_RESPONSE",
        message: "Invalid server response",
        requestId: responseRequestID,
      });
    }

    if (!isCanonicalApiErrorEnvelope(errorPayload)) {
      return makeError({
        kind: "invalid_response",
        status: response.status,
        code: "INVALID_RESPONSE",
        message: "Invalid server response",
        requestId: responseRequestID,
      });
    }

    const requestId = errorPayload.error.request_id || responseRequestID;
    if (response.status === 401) {
      return makeError({
        kind: "unauthorized",
        status: response.status,
        code: errorPayload.error.code,
        message: errorPayload.error.message,
        requestId,
      });
    }

    if (response.status === 403) {
      return makeError({
        kind: "forbidden",
        status: response.status,
        code: errorPayload.error.code,
        message: errorPayload.error.message,
        requestId,
      });
    }

    return makeError({
      kind: "api_error",
      status: response.status,
      code: errorPayload.error.code,
      message: errorPayload.error.message,
      requestId,
    });
  }

  return {
    requestJson,
  };
}
