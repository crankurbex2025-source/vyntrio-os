export type PublicSettingsDto = {
  instance: {
    name: string;
    version: string;
  };
  api: {
    environment: string;
  };
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

export function parsePublicSettingsDto(payload: unknown): PublicSettingsDto | null {
  if (!isPlainRecord(payload) || !hasExactKeys(payload, ["instance", "api"])) {
    return null;
  }

  const instance = payload.instance;
  const api = payload.api;

  if (!isPlainRecord(instance) || !hasExactKeys(instance, ["name", "version"])) {
    return null;
  }
  if (!isPlainRecord(api) || !hasExactKeys(api, ["environment"])) {
    return null;
  }

  if (
    typeof instance.name !== "string" ||
    typeof instance.version !== "string" ||
    typeof api.environment !== "string"
  ) {
    return null;
  }

  return {
    instance: {
      name: instance.name,
      version: instance.version,
    },
    api: {
      environment: api.environment,
    },
  };
}
