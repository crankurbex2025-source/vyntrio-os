export type OverviewDto = {
  instance: {
    name: string;
    version: string;
    commit: string;
  };
  api: {
    environment: string;
  };
  service: {
    status: "running";
  };
  readiness: {
    status: "ready" | "not_ready";
    database: "ok" | "error";
  };
  collected_at: string;
};

function isPlainRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function hasExactKeys(record: Record<string, unknown>, keys: string[]): boolean {
  const actual = Object.keys(record).sort();
  const expected = [...keys].sort();
  if (actual.length !== expected.length) {
    return false;
  }
  return expected.every((key, index) => key === actual[index]);
}

export function parseOverviewDto(payload: unknown): OverviewDto | null {
  if (
    !isPlainRecord(payload) ||
    !hasExactKeys(payload, ["instance", "api", "service", "readiness", "collected_at"])
  ) {
    return null;
  }

  const instance = payload.instance;
  const api = payload.api;
  const service = payload.service;
  const readiness = payload.readiness;
  const collectedAt = payload.collected_at;

  if (!isPlainRecord(instance) || !hasExactKeys(instance, ["name", "version", "commit"])) {
    return null;
  }
  if (!isPlainRecord(api) || !hasExactKeys(api, ["environment"])) {
    return null;
  }
  if (!isPlainRecord(service) || !hasExactKeys(service, ["status"])) {
    return null;
  }
  if (!isPlainRecord(readiness) || !hasExactKeys(readiness, ["status", "database"])) {
    return null;
  }

  if (
    typeof instance.name !== "string" ||
    typeof instance.version !== "string" ||
    typeof instance.commit !== "string" ||
    typeof api.environment !== "string" ||
    service.status !== "running" ||
    typeof collectedAt !== "string" ||
    (readiness.status !== "ready" && readiness.status !== "not_ready") ||
    (readiness.database !== "ok" && readiness.database !== "error")
  ) {
    return null;
  }

  return {
    instance: {
      name: instance.name,
      version: instance.version,
      commit: instance.commit,
    },
    api: {
      environment: api.environment,
    },
    service: {
      status: "running",
    },
    readiness: {
      status: readiness.status,
      database: readiness.database,
    },
    collected_at: collectedAt,
  };
}

export function formatOverviewCollectedAt(value: string): string {
  const parsed = Date.parse(value);
  if (Number.isNaN(parsed)) {
    return value;
  }
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "medium",
  }).format(parsed);
}
