export function formatSupportStatus(value: string | undefined): string {
  switch (value) {
    case "engineering_media_early_access":
      return "Engineering media / early access (testing only)";
    default:
      return value ?? "—";
  }
}

export function formatPublicationStatus(
  status: string | undefined,
  labels: {
    notBuilt: string;
    localStaging: string;
    unavailable: string;
  }
): string {
  switch (status) {
    case "local_staging":
      return labels.localStaging;
    case "unavailable":
      return labels.unavailable;
    case "not_built":
      return labels.notBuilt;
    default:
      return labels.notBuilt;
  }
}
