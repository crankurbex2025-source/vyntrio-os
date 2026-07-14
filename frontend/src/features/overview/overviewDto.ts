export type MetricStatus = "ok" | "unavailable";

export type HostCpuDto = {
  status: MetricStatus;
  logical_cores?: number;
  load_1m?: number;
};

export type HostMemoryDto = {
  status: MetricStatus;
  total_bytes?: number;
  available_bytes?: number;
  used_bytes?: number;
};

export type HostFilesystemDto = {
  id: "state";
  status: MetricStatus;
  total_bytes?: number;
  available_bytes?: number;
  used_bytes?: number;
  fs_type?: "ext4" | "xfs" | "btrfs" | "tmpfs" | "other";
};

export type HostDto = {
  cpu: HostCpuDto;
  memory: HostMemoryDto;
  filesystems: [HostFilesystemDto];
};

export type BackupStatus = "never_run" | "succeeded" | "failed" | "unavailable";

export type BackupFailure = "artifact" | "restart" | "health" | "readiness" | "internal";

export type BackupDto = {
  status: BackupStatus;
  completed_at?: string;
  ever_succeeded?: boolean;
  failure?: BackupFailure;
};

export type NetworkStatus = "available" | "unknown" | "unavailable";

export type NetworkDto = {
  status: NetworkStatus;
};

export type SoftwareStatus = "ok" | "unavailable";

export type ReleaseChannel = "development" | "production" | "unknown";

export type SoftwareDto = {
  status: SoftwareStatus;
  version?: string;
  commit?: string;
  channel?: ReleaseChannel;
};

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
  host: HostDto;
  backup: BackupDto;
  network: NetworkDto;
  software: SoftwareDto;
  collected_at: string;
};

const BACKUP_FAILURES = new Set(["artifact", "restart", "health", "readiness", "internal"]);

const FS_TYPES = new Set(["ext4", "xfs", "btrfs", "tmpfs", "other"]);

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

function isPositiveInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value > 0;
}

function isNonNegativeInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value >= 0;
}

function isNonNegativeNumber(value: unknown): value is number {
  return typeof value === "number" && Number.isFinite(value) && value >= 0;
}

function parseMetricStatus(value: unknown): MetricStatus | null {
  if (value === "ok" || value === "unavailable") {
    return value;
  }
  return null;
}

function parseHostCpu(value: unknown): HostCpuDto | null {
  if (!isPlainRecord(value) || !("status" in value)) {
    return null;
  }
  const status = parseMetricStatus(value.status);
  if (!status) {
    return null;
  }
  if (status === "unavailable") {
    return hasExactKeys(value, ["status"]) ? { status } : null;
  }
  if (
    !hasExactKeys(value, ["status", "logical_cores", "load_1m"]) ||
    !isPositiveInteger(value.logical_cores) ||
    !isNonNegativeNumber(value.load_1m)
  ) {
    return null;
  }
  return {
    status,
    logical_cores: value.logical_cores,
    load_1m: value.load_1m,
  };
}

function parseHostMemory(value: unknown): HostMemoryDto | null {
  if (!isPlainRecord(value) || !("status" in value)) {
    return null;
  }
  const status = parseMetricStatus(value.status);
  if (!status) {
    return null;
  }
  if (status === "unavailable") {
    return hasExactKeys(value, ["status"]) ? { status } : null;
  }
  if (
    !hasExactKeys(value, ["status", "total_bytes", "available_bytes", "used_bytes"]) ||
    !isPositiveInteger(value.total_bytes) ||
    !isNonNegativeInteger(value.available_bytes) ||
    !isNonNegativeInteger(value.used_bytes) ||
    value.available_bytes > value.total_bytes ||
    value.used_bytes !== value.total_bytes - value.available_bytes
  ) {
    return null;
  }
  return {
    status,
    total_bytes: value.total_bytes,
    available_bytes: value.available_bytes,
    used_bytes: value.used_bytes,
  };
}

function parseHostFilesystem(value: unknown): HostFilesystemDto | null {
  if (!isPlainRecord(value) || !("status" in value) || value.id !== "state") {
    return null;
  }
  const status = parseMetricStatus(value.status);
  if (!status) {
    return null;
  }
  if (status === "unavailable") {
    return hasExactKeys(value, ["id", "status"]) ? { id: "state", status } : null;
  }
  const keys = ["id", "status", "total_bytes", "available_bytes", "used_bytes"];
  const hasFsType = "fs_type" in value;
  if (hasFsType) {
    keys.push("fs_type");
  }
  if (
    !hasExactKeys(value, keys) ||
    !isPositiveInteger(value.total_bytes) ||
    !isNonNegativeInteger(value.available_bytes) ||
    !isNonNegativeInteger(value.used_bytes) ||
    value.available_bytes > value.total_bytes ||
    value.used_bytes !== value.total_bytes - value.available_bytes
  ) {
    return null;
  }
  let fsType: HostFilesystemDto["fs_type"] | undefined;
  if (hasFsType) {
    if (typeof value.fs_type !== "string" || !FS_TYPES.has(value.fs_type)) {
      return null;
    }
    fsType = value.fs_type as HostFilesystemDto["fs_type"];
  }
  return {
    id: "state",
    status,
    total_bytes: value.total_bytes,
    available_bytes: value.available_bytes,
    used_bytes: value.used_bytes,
    ...(fsType ? { fs_type: fsType } : {}),
  };
}

function parseHost(value: unknown): HostDto | null {
  if (!isPlainRecord(value) || !hasExactKeys(value, ["cpu", "memory", "filesystems"])) {
    return null;
  }
  const cpu = parseHostCpu(value.cpu);
  const memory = parseHostMemory(value.memory);
  if (!cpu || !memory || !Array.isArray(value.filesystems) || value.filesystems.length !== 1) {
    return null;
  }
  const filesystem = parseHostFilesystem(value.filesystems[0]);
  if (!filesystem) {
    return null;
  }
  return {
    cpu,
    memory,
    filesystems: [filesystem],
  };
}

function parseBackup(value: unknown): BackupDto | null {
  if (!isPlainRecord(value) || !("status" in value)) {
    return null;
  }
  const status = value.status;
  if (
    status !== "never_run" &&
    status !== "succeeded" &&
    status !== "failed" &&
    status !== "unavailable"
  ) {
    return null;
  }
  if (status === "unavailable") {
    return hasExactKeys(value, ["status"]) ? { status } : null;
  }
  if (status === "never_run") {
    if (
      !hasExactKeys(value, ["status", "ever_succeeded"]) ||
      value.ever_succeeded !== false
    ) {
      return null;
    }
    return { status, ever_succeeded: false };
  }
  if (status === "succeeded") {
    if (
      !hasExactKeys(value, ["status", "completed_at", "ever_succeeded"]) ||
      typeof value.completed_at !== "string" ||
      value.ever_succeeded !== true
    ) {
      return null;
    }
    return {
      status,
      completed_at: value.completed_at,
      ever_succeeded: true,
    };
  }
  if (
    !hasExactKeys(value, ["status", "completed_at", "ever_succeeded", "failure"]) ||
    typeof value.completed_at !== "string" ||
    typeof value.ever_succeeded !== "boolean" ||
    typeof value.failure !== "string" ||
    !BACKUP_FAILURES.has(value.failure)
  ) {
    return null;
  }
  return {
    status,
    completed_at: value.completed_at,
    ever_succeeded: value.ever_succeeded,
    failure: value.failure as BackupFailure,
  };
}

function parseNetwork(value: unknown): NetworkDto | null {
  if (!isPlainRecord(value) || !hasExactKeys(value, ["status"])) {
    return null;
  }
  const status = value.status;
  if (status !== "available" && status !== "unknown" && status !== "unavailable") {
    return null;
  }
  return { status };
}

function parseSoftware(value: unknown): SoftwareDto | null {
  if (!isPlainRecord(value) || !("status" in value)) {
    return null;
  }
  const status = value.status;
  if (status !== "ok" && status !== "unavailable") {
    return null;
  }
  if (status === "unavailable") {
    return hasExactKeys(value, ["status"]) ? { status } : null;
  }
  const hasCommit = typeof value.commit === "string";
  const expectedKeys = hasCommit
    ? ["status", "version", "commit", "channel"]
    : ["status", "version", "channel"];
  if (!hasExactKeys(value, expectedKeys)) {
    return null;
  }
  const channel = value.channel;
  if (
    typeof value.version !== "string" ||
    (channel !== "development" && channel !== "production" && channel !== "unknown")
  ) {
    return null;
  }
  if (typeof value.commit === "string") {
    return {
      status,
      version: value.version,
      commit: value.commit,
      channel,
    };
  }
  return {
    status,
    version: value.version,
    channel,
  };
}

export function parseOverviewDto(payload: unknown): OverviewDto | null {
  if (
    !isPlainRecord(payload) ||
    !hasExactKeys(payload, [
      "instance",
      "api",
      "service",
      "readiness",
      "host",
      "backup",
      "network",
      "software",
      "collected_at",
    ])
  ) {
    return null;
  }

  const instance = payload.instance;
  const api = payload.api;
  const service = payload.service;
  const readiness = payload.readiness;
  const host = parseHost(payload.host);
  const backup = parseBackup(payload.backup);
  const network = parseNetwork(payload.network);
  const software = parseSoftware(payload.software);
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
  if (!host) {
    return null;
  }
  if (!backup) {
    return null;
  }
  if (!network) {
    return null;
  }
  if (!software) {
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
    host,
    backup,
    network,
    software,
    collected_at: collectedAt,
  };
}

export function formatBackupFailureDetail(failure: BackupFailure): string {
  switch (failure) {
    case "artifact":
      return "Backup artifact creation or validation did not complete.";
    case "restart":
      return "The API service did not restart successfully afterward.";
    case "health":
      return "Local health checks failed after restart.";
    case "readiness":
      return "The database was not ready after restart.";
    default:
      return "The backup command did not complete successfully.";
  }
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

export function formatMetricBytes(value: number): string {
  if (!Number.isFinite(value) || value < 0) {
    return "Unavailable";
  }
  const units = ["B", "KB", "MB", "GB", "TB"];
  let size = value;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex += 1;
  }
  const precision =
    size >= 10 || (Number.isInteger(size) && unitIndex > 0) ? 0 : size >= 1 ? 1 : 0;
  const formatted = precision === 0 ? Math.round(size).toString() : size.toFixed(precision);
  return `${formatted} ${units[unitIndex]}`;
}
