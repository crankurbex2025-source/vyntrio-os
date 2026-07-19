export type ApplianceNavItem = {
  id: string;
  label: string;
  to: string;
  /** When true, shown in nav but marked as not available yet. */
  planned?: boolean;
};

/**
 * Single source of truth for post-boot appliance navigation.
 * Apps/VMs stay out of primary nav until runtimes exist (capability-gated later).
 */
export function buildApplianceNavItems(): ApplianceNavItem[] {
  return [
    { id: "dashboard", label: "Dashboard", to: "/app" },
    { id: "storage", label: "Storage", to: "/app/storage" },
    { id: "shares", label: "Shares", to: "/app/shares" },
    { id: "users", label: "Users", to: "/app/users", planned: true },
    { id: "settings", label: "Settings", to: "/app/settings" },
    { id: "tools", label: "Tools", to: "/app/tools", planned: true },
  ];
}

export function applianceSectionFromPath(
  pathname: string
): ApplianceNavItem["id"] {
  if (pathname === "/app" || pathname === "/app/") {
    return "dashboard";
  }
  if (pathname.startsWith("/app/storage")) {
    return "storage";
  }
  if (pathname.startsWith("/app/shares")) {
    return "shares";
  }
  if (pathname.startsWith("/app/users")) {
    return "users";
  }
  if (pathname.startsWith("/app/settings")) {
    return "settings";
  }
  if (pathname.startsWith("/app/tools")) {
    return "tools";
  }
  return "dashboard";
}
