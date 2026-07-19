export type InstallMediaPublicationStatus = "not_built" | "local_staging" | "unavailable";

export type InstallMediaWriterArtifact = {
  platform: string;
  arch: string;
  name: string;
  kind: string;
  size_bytes?: number;
  sha256?: string;
  download_available: boolean;
  download_path?: string;
};

export type InstallMediaDto = {
  publication_status: InstallMediaPublicationStatus;
  generated_at?: string;
  release: {
    version: string;
    channel?: string;
    build_id?: string;
  };
  primary_artifact: {
    name: string;
    format: string;
    firmware_boot_mode: string;
    bios_support?: boolean;
    uefi_support?: boolean;
    dual_mode?: boolean;
    secure_boot?: string;
    media_role?: string;
    size_bytes?: number;
    sha256?: string;
    download_available: boolean;
    download_path?: string;
    manifest_path?: string;
  };
  image_versions?: Array<{
    version: string;
    build_id?: string;
    channel?: string;
    generated_at?: string;
    name: string;
    format: string;
    size_bytes?: number;
    sha256?: string;
    firmware_boot_mode: string;
    bios_support?: boolean;
    uefi_support?: boolean;
    dual_mode?: boolean;
    secure_boot?: string;
    latest?: boolean;
    media_role?: string;
  }>;
  build_target: string;
  stage_target: string;
  verify_command?: string;
  support_status?: string;
  writer?: {
    name: string;
    kind: string;
    platforms: string[];
    binary_name: string;
    build_target: string;
    package_target?: string;
    documentation_path: string;
    requires_elevation: boolean;
    native_gui: boolean;
    gui_available?: boolean;
    gui_kind?: string;
    artifacts: InstallMediaWriterArtifact[];
  };
  limitations: string[];
};

const PUBLICATION_STATUSES = new Set<InstallMediaPublicationStatus>([
  "not_built",
  "local_staging",
  "unavailable",
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

function isNonNegativeInteger(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && value >= 0;
}

export function parseInstallMediaDto(payload: unknown): InstallMediaDto | null {
  if (!isPlainRecord(payload) || !("publication_status" in payload)) {
    return null;
  }
  if (!PUBLICATION_STATUSES.has(payload.publication_status as InstallMediaPublicationStatus)) {
    return null;
  }
  if (!isPlainRecord(payload.release) || typeof payload.release.version !== "string") {
    return null;
  }
  if (!isPlainRecord(payload.primary_artifact) || typeof payload.primary_artifact.name !== "string") {
    return null;
  }
  if (typeof payload.build_target !== "string" || typeof payload.stage_target !== "string") {
    return null;
  }
  if (!Array.isArray(payload.limitations)) {
    return null;
  }
  for (const item of payload.limitations) {
    if (typeof item !== "string") {
      return null;
    }
  }

  const artifact = payload.primary_artifact;
  if (
    typeof artifact.format !== "string" ||
    typeof artifact.firmware_boot_mode !== "string" ||
    typeof artifact.download_available !== "boolean"
  ) {
    return null;
  }
  if (artifact.size_bytes !== undefined && !isNonNegativeInteger(artifact.size_bytes)) {
    return null;
  }
  if (artifact.sha256 !== undefined && typeof artifact.sha256 !== "string") {
    return null;
  }
  if (artifact.download_path !== undefined && typeof artifact.download_path !== "string") {
    return null;
  }
  if (artifact.manifest_path !== undefined && typeof artifact.manifest_path !== "string") {
    return null;
  }
  if (payload.generated_at !== undefined && typeof payload.generated_at !== "string") {
    return null;
  }
  if (payload.verify_command !== undefined && typeof payload.verify_command !== "string") {
    return null;
  }
  if (payload.support_status !== undefined && typeof payload.support_status !== "string") {
    return null;
  }
  if (payload.writer !== undefined) {
    if (!isPlainRecord(payload.writer) || typeof payload.writer.name !== "string") {
      return null;
    }
  }
  if (payload.release.channel !== undefined && typeof payload.release.channel !== "string") {
    return null;
  }

  return {
    publication_status: payload.publication_status as InstallMediaPublicationStatus,
    generated_at: payload.generated_at as string | undefined,
    release: {
      version: payload.release.version,
      channel: payload.release.channel as string | undefined,
      build_id: payload.release.build_id as string | undefined,
    },
    primary_artifact: {
      name: artifact.name,
      format: artifact.format,
      firmware_boot_mode: artifact.firmware_boot_mode,
      bios_support: artifact.bios_support === true,
      uefi_support: artifact.uefi_support === true,
      dual_mode: artifact.dual_mode === true,
      secure_boot:
        typeof artifact.secure_boot === "string" ? artifact.secure_boot : "unsupported",
      media_role: typeof artifact.media_role === "string" ? artifact.media_role : "appliance",
      size_bytes: artifact.size_bytes as number | undefined,
      sha256: artifact.sha256 as string | undefined,
      download_available: artifact.download_available,
      download_path: artifact.download_path as string | undefined,
      manifest_path: artifact.manifest_path as string | undefined,
    },
    image_versions: Array.isArray(payload.image_versions)
      ? (payload.image_versions as InstallMediaDto["image_versions"])
      : undefined,
    build_target: payload.build_target,
    stage_target: payload.stage_target,
    verify_command: payload.verify_command as string | undefined,
    support_status: payload.support_status as string | undefined,
    writer: payload.writer
      ? {
          name: (payload.writer as Record<string, unknown>).name as string,
          kind: ((payload.writer as Record<string, unknown>).kind as string) ?? "",
          platforms: Array.isArray((payload.writer as Record<string, unknown>).platforms)
            ? ((payload.writer as Record<string, unknown>).platforms as unknown[]).filter(
                (item): item is string => typeof item === "string"
              )
            : [],
          binary_name: ((payload.writer as Record<string, unknown>).binary_name as string) ?? "",
          build_target: ((payload.writer as Record<string, unknown>).build_target as string) ?? "",
          package_target: (payload.writer as Record<string, unknown>).package_target as
            | string
            | undefined,
          documentation_path:
            ((payload.writer as Record<string, unknown>).documentation_path as string) ?? "",
          requires_elevation:
            (payload.writer as Record<string, unknown>).requires_elevation === true,
          native_gui: (payload.writer as Record<string, unknown>).native_gui === true,
          gui_available: (payload.writer as Record<string, unknown>).gui_available === true,
          gui_kind: (payload.writer as Record<string, unknown>).gui_kind as string | undefined,
          artifacts: parseWriterArtifacts(
            (payload.writer as Record<string, unknown>).artifacts
          ),
        }
      : undefined,
    limitations: payload.limitations as string[],
  };
}

function parseWriterArtifacts(value: unknown): InstallMediaWriterArtifact[] {
  if (!Array.isArray(value)) {
    return [];
  }
  const artifacts: InstallMediaWriterArtifact[] = [];
  for (const item of value) {
    if (!isPlainRecord(item)) {
      continue;
    }
    if (typeof item.platform !== "string" || typeof item.name !== "string") {
      continue;
    }
    artifacts.push({
      platform: item.platform,
      arch: typeof item.arch === "string" ? item.arch : "",
      name: item.name,
      kind: typeof item.kind === "string" ? item.kind : "",
      size_bytes: isNonNegativeInteger(item.size_bytes) ? item.size_bytes : undefined,
      sha256: typeof item.sha256 === "string" ? item.sha256 : undefined,
      download_available: item.download_available === true,
      download_path: typeof item.download_path === "string" ? item.download_path : undefined,
    });
  }
  return artifacts;
}

export function formatInstallMediaBytes(value: number | undefined): string {
  if (value === undefined || !Number.isFinite(value) || value < 0) {
    return "—";
  }
  const units = ["B", "KB", "MB", "GB"];
  let size = value;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex += 1;
  }
  const precision = size >= 10 || unitIndex === 0 ? 0 : 1;
  const formatted = precision === 0 ? Math.round(size).toString() : size.toFixed(precision);
  return `${formatted} ${units[unitIndex]}`;
}
