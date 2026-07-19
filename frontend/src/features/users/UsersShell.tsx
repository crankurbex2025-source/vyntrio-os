import { ApplianceDataTable } from "../../surfaces/appliance/ui/ApplianceDataTable";
import { AppliancePageHeader } from "../../surfaces/appliance/ui/AppliancePageHeader";
import { AppliancePanel } from "../../surfaces/appliance/ui/AppliancePanel";

export function UsersShell() {
  return (
    <div className="vyn-ops-page">
      <AppliancePageHeader
        title="Users"
        status="Owner session only · multi-user management is not available yet"
      />

      <AppliancePanel
        title="Accounts"
        planned
        note="This appliance authenticates a single Owner session. Local users, groups, and share ACLs are not implemented."
      >
        <ApplianceDataTable
          caption="User accounts"
          columns={[
            { key: "account", header: "Account" },
            { key: "role", header: "Role" },
            { key: "status", header: "Status" },
          ]}
          rows={[
            {
              account: "Owner",
              role: "Owner",
              status: "Signed-in session (CRUD UI not shipping)",
            },
          ]}
        />
      </AppliancePanel>
    </div>
  );
}
