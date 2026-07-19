import { formatMetricBytes, formatOverviewCollectedAt } from "../overview/overviewDto";

export type InventoryStatus = "ok" | "unavailable";

export type DiskDeviceStatus = "ok" | "unavailable";

export type DiskEligibility = "eligible" | "excluded" | "unknown";

export type DiskExclusionReason =
  | "root_disk"
  | "state_filesystem"
  | "removable"
  | "read_only"
  | "mounted_in_use"
  | "unsupported_filesystem"
  | "install_media"
  | "virtual_device"
  | "ambiguous_identity";

export type DiskDeviceDto = {
  id: string;
  status: DiskDeviceStatus;
  size_bytes?: number;
  rotational?: boolean;
  removable?: boolean;
  eligibility: DiskEligibility;
  reasons?: DiskExclusionReason[];
};

export type StorageDisksDto = {
  collected_at: string;
  status: InventoryStatus;
  disks: DiskDeviceDto[];
};

export type LayoutStatus = "ok" | "unavailable";

export type PoolManagementAvailability = "declared_pools" | "not_available";

export type ShareManagementAvailability = "planned_shares" | "not_available";

export type DatasetDto = {
  id: string;
  name: string;
  path_intent: string;
  status: string;
  created_at: string;
};

export type PoolDto = {
  id: string;
  name: string;
  status: string;
  disk_ids: string[];
  disk_format_state: string;
  datasets: DatasetDto[];
  created_at: string;
  updated_at: string;
};

export type ShareDto = {
  id: string;
  name: string;
  pool_id: string;
  dataset_id?: string;
  protocol: string;
  status: string;
  created_at: string;
};

export type StoragePoolsDto = {
  collected_at: string;
  status: LayoutStatus;
  inventory_status: InventoryStatus;
  pools: PoolDto[];
  pool_management: PoolManagementAvailability;
  mutation_available: boolean;
  disk_format_applied?: boolean;
  note?: string;
};

export type StorageSharesDto = {
  collected_at: string;
  status: LayoutStatus;
  inventory_status: InventoryStatus;
  shares: ShareDto[];
  share_management: ShareManagementAvailability;
  protocol_support: "not_available" | string;
  mutation_available: boolean;
  note?: string;
};

export type StorageLayoutDto = {
  disks: StorageDisksDto;
  pools: StoragePoolsDto;
  shares: StorageSharesDto;
};

const INVENTORY_STATUSES = new Set<InventoryStatus>(["ok", "unavailable"]);
const LAYOUT_STATUSES = new Set<LayoutStatus>(["ok", "unavailable"]);
const DEVICE_STATUSES = new Set<DiskDeviceStatus>(["ok", "unavailable"]);
const ELIGIBILITIES = new Set<DiskEligibility>(["eligible", "excluded", "unknown"]);
const EXCLUSION_REASONS = new Set<DiskExclusionReason>([
  "root_disk",
  "state_filesystem",
  "removable",
  "read_only",
  "mounted_in_use",
  "unsupported_filesystem",
  "install_media",
  "virtual_device",
  "ambiguous_identity",
]);

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

function parseReasons(value: unknown): DiskExclusionReason[] | undefined | null {
  if (value === undefined) {
    return undefined;
  }
  if (!Array.isArray(value)) {
    return null;
  }
  const reasons: DiskExclusionReason[] = [];
  for (const item of value) {
    if (typeof item !== "string" || !EXCLUSION_REASONS.has(item as DiskExclusionReason)) {
      return null;
    }
    reasons.push(item as DiskExclusionReason);
  }
  return reasons;
}

function parseDiskDevice(value: unknown): DiskDeviceDto | null {
  if (!isPlainRecord(value) || !("status" in value) || !("eligibility" in value) || !("id" in value)) {
    return null;
  }
  if (typeof value.id !== "string" || value.id.length === 0) {
    return null;
  }
  if (!DEVICE_STATUSES.has(value.status as DiskDeviceStatus)) {
    return null;
  }
  if (!ELIGIBILITIES.has(value.eligibility as DiskEligibility)) {
    return null;
  }

  const status = value.status as DiskDeviceStatus;
  const eligibility = value.eligibility as DiskEligibility;
  const reasons = parseReasons(value.reasons);
  if (value.reasons !== undefined && reasons === null) {
    return null;
  }

  const baseKeys = ["id", "status", "eligibility"];
  const optionalKeys: string[] = [];
  if (value.size_bytes !== undefined) {
    optionalKeys.push("size_bytes");
  }
  if (value.rotational !== undefined) {
    optionalKeys.push("rotational");
  }
  if (value.removable !== undefined) {
    optionalKeys.push("removable");
  }
  if (value.reasons !== undefined) {
    optionalKeys.push("reasons");
  }
  if (!hasExactKeys(value, [...baseKeys, ...optionalKeys])) {
    return null;
  }

  if (value.size_bytes !== undefined && !isPositiveInteger(value.size_bytes)) {
    return null;
  }
  if (value.rotational !== undefined && typeof value.rotational !== "boolean") {
    return null;
  }
  if (value.removable !== undefined && typeof value.removable !== "boolean") {
    return null;
  }

  return {
    id: value.id,
    status,
    eligibility,
    ...(value.size_bytes !== undefined ? { size_bytes: value.size_bytes } : {}),
    ...(value.rotational !== undefined ? { rotational: value.rotational } : {}),
    ...(value.removable !== undefined ? { removable: value.removable } : {}),
    ...(reasons && reasons.length > 0 ? { reasons } : {}),
  };
}

export function parseStorageDisksDto(payload: unknown): StorageDisksDto | null {
  if (!isPlainRecord(payload) || !hasExactKeys(payload, ["collected_at", "status", "disks"])) {
    return null;
  }
  if (typeof payload.collected_at !== "string" || payload.collected_at.length === 0) {
    return null;
  }
  if (!INVENTORY_STATUSES.has(payload.status as InventoryStatus)) {
    return null;
  }
  if (!Array.isArray(payload.disks)) {
    return null;
  }

  const disks: DiskDeviceDto[] = [];
  for (const item of payload.disks) {
    const disk = parseDiskDevice(item);
    if (!disk) {
      return null;
    }
    disks.push(disk);
  }

  return {
    collected_at: payload.collected_at,
    status: payload.status as InventoryStatus,
    disks,
  };
}

function parseDataset(value: unknown): DatasetDto | null {
  if (!isPlainRecord(value)) {
    return null;
  }
  if (
    typeof value.id !== "string" ||
    typeof value.name !== "string" ||
    typeof value.path_intent !== "string" ||
    typeof value.status !== "string" ||
    typeof value.created_at !== "string"
  ) {
    return null;
  }
  return {
    id: value.id,
    name: value.name,
    path_intent: value.path_intent,
    status: value.status,
    created_at: value.created_at,
  };
}

function parsePool(value: unknown): PoolDto | null {
  if (!isPlainRecord(value)) {
    return null;
  }
  if (
    typeof value.id !== "string" ||
    typeof value.name !== "string" ||
    typeof value.status !== "string" ||
    typeof value.disk_format_state !== "string" ||
    typeof value.created_at !== "string" ||
    typeof value.updated_at !== "string" ||
    !Array.isArray(value.disk_ids) ||
    !Array.isArray(value.datasets)
  ) {
    return null;
  }
  const diskIds: string[] = [];
  for (const id of value.disk_ids) {
    if (typeof id !== "string") {
      return null;
    }
    diskIds.push(id);
  }
  const datasets: DatasetDto[] = [];
  for (const item of value.datasets) {
    const ds = parseDataset(item);
    if (!ds) {
      return null;
    }
    datasets.push(ds);
  }
  return {
    id: value.id,
    name: value.name,
    status: value.status,
    disk_ids: diskIds,
    disk_format_state: value.disk_format_state,
    datasets,
    created_at: value.created_at,
    updated_at: value.updated_at,
  };
}

function parseShare(value: unknown): ShareDto | null {
  if (!isPlainRecord(value)) {
    return null;
  }
  if (
    typeof value.id !== "string" ||
    typeof value.name !== "string" ||
    typeof value.pool_id !== "string" ||
    typeof value.protocol !== "string" ||
    typeof value.status !== "string" ||
    typeof value.created_at !== "string"
  ) {
    return null;
  }
  if (value.dataset_id !== undefined && typeof value.dataset_id !== "string") {
    return null;
  }
  return {
    id: value.id,
    name: value.name,
    pool_id: value.pool_id,
    ...(value.dataset_id ? { dataset_id: value.dataset_id as string } : {}),
    protocol: value.protocol,
    status: value.status,
    created_at: value.created_at,
  };
}

export function parseStoragePoolsDto(payload: unknown): StoragePoolsDto | null {
  if (!isPlainRecord(payload)) {
    return null;
  }
  if (
    typeof payload.collected_at !== "string" ||
    typeof payload.status !== "string" ||
    typeof payload.inventory_status !== "string" ||
    typeof payload.pool_management !== "string" ||
    typeof payload.mutation_available !== "boolean" ||
    !Array.isArray(payload.pools)
  ) {
    return null;
  }
  if (!LAYOUT_STATUSES.has(payload.status as LayoutStatus)) {
    return null;
  }
  if (!INVENTORY_STATUSES.has(payload.inventory_status as InventoryStatus)) {
    return null;
  }
  if (payload.pool_management !== "declared_pools" && payload.pool_management !== "not_available") {
    return null;
  }
  const pools: PoolDto[] = [];
  for (const item of payload.pools) {
    const pool = parsePool(item);
    if (!pool) {
      return null;
    }
    pools.push(pool);
  }
  return {
    collected_at: payload.collected_at,
    status: payload.status as LayoutStatus,
    inventory_status: payload.inventory_status as InventoryStatus,
    pools,
    pool_management: payload.pool_management as PoolManagementAvailability,
    mutation_available: payload.mutation_available,
    ...(typeof payload.disk_format_applied === "boolean"
      ? { disk_format_applied: payload.disk_format_applied }
      : {}),
    ...(typeof payload.note === "string" ? { note: payload.note } : {}),
  };
}

export function parseStorageSharesDto(payload: unknown): StorageSharesDto | null {
  if (!isPlainRecord(payload)) {
    return null;
  }
  if (
    typeof payload.collected_at !== "string" ||
    typeof payload.status !== "string" ||
    typeof payload.inventory_status !== "string" ||
    typeof payload.share_management !== "string" ||
    typeof payload.protocol_support !== "string" ||
    typeof payload.mutation_available !== "boolean" ||
    !Array.isArray(payload.shares)
  ) {
    return null;
  }
  if (!LAYOUT_STATUSES.has(payload.status as LayoutStatus)) {
    return null;
  }
  if (!INVENTORY_STATUSES.has(payload.inventory_status as InventoryStatus)) {
    return null;
  }
  if (payload.share_management !== "planned_shares" && payload.share_management !== "not_available") {
    return null;
  }
  const shares: ShareDto[] = [];
  for (const item of payload.shares) {
    const share = parseShare(item);
    if (!share) {
      return null;
    }
    shares.push(share);
  }
  return {
    collected_at: payload.collected_at,
    status: payload.status as LayoutStatus,
    inventory_status: payload.inventory_status as InventoryStatus,
    shares,
    share_management: payload.share_management as ShareManagementAvailability,
    protocol_support: payload.protocol_support,
    mutation_available: payload.mutation_available,
    ...(typeof payload.note === "string" ? { note: payload.note } : {}),
  };
}

export function parseStorageLayoutDto(payload: {
  disks: unknown;
  pools: unknown;
  shares: unknown;
}): StorageLayoutDto | null {
  const disks = parseStorageDisksDto(payload.disks);
  const pools = parseStoragePoolsDto(payload.pools);
  const shares = parseStorageSharesDto(payload.shares);
  if (!disks || !pools || !shares) {
    return null;
  }
  return { disks, pools, shares };
}

export function formatDiskSize(sizeBytes: number | undefined): string {
  if (sizeBytes === undefined) {
    return "Unknown size";
  }
  return formatMetricBytes(sizeBytes);
}

export function formatDiskCollectedAt(collectedAt: string): string {
  return formatOverviewCollectedAt(collectedAt);
}

export function formatEligibilityLabel(eligibility: DiskEligibility): string {
  switch (eligibility) {
    case "eligible":
      return "Eligible candidate";
    case "excluded":
      return "Excluded";
    case "unknown":
      return "Unknown";
    default:
      return "Unknown";
  }
}

export function formatExclusionReason(reason: DiskExclusionReason): string {
  switch (reason) {
    case "root_disk":
      return "Carries the root filesystem";
    case "state_filesystem":
      return "Carries appliance state storage";
    case "removable":
      return "Removable media";
    case "read_only":
      return "Read-only block device";
    case "mounted_in_use":
      return "Mounted and in use";
    case "unsupported_filesystem":
      return "Unsupported filesystem for pools";
    case "install_media":
      return "Install or optical media";
    case "virtual_device":
      return "Virtual block device";
    case "ambiguous_identity":
      return "Ambiguous or incomplete identity";
    default:
      return reason;
  }
}

export function countEligibleDisks(disks: DiskDeviceDto[]): number {
  return disks.filter((disk) => disk.eligibility === "eligible").length;
}

export function eligibleDisks(disks: DiskDeviceDto[]): DiskDeviceDto[] {
  return disks.filter((disk) => disk.eligibility === "eligible");
}
