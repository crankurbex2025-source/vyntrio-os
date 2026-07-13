package identity_test

import (
	"testing"

	"github.com/crankurbex2025-source/vyntrio-os/internal/domain/identity"
)

// expectedMatrix is the normative ADR-0004 v1 role/permission matrix.
var expectedMatrix = map[identity.Role]map[identity.Permission]bool{
	identity.RoleOwner: {
		identity.PermissionSystemHealth:       true,
		identity.PermissionAuthSession:        true,
		identity.PermissionUsersRead:          true,
		identity.PermissionUsersWrite:         true,
		identity.PermissionRolesAssign:        true,
		identity.PermissionRolesAssignOwner:   true,
		identity.PermissionSettingsRead:       true,
		identity.PermissionSettingsWrite:      true,
		identity.PermissionSettingsAdminRead:  true,
		identity.PermissionSettingsAdminWrite: true,
		identity.PermissionStorageRead:        true,
		identity.PermissionStorageWrite:       true,
		identity.PermissionContainersRead:     true,
		identity.PermissionContainersWrite:    true,
		identity.PermissionVMsRead:            true,
		identity.PermissionVMsWrite:           true,
		identity.PermissionNetworkRead:        true,
		identity.PermissionNetworkWrite:       true,
		identity.PermissionUpdatesRead:        true,
		identity.PermissionUpdatesApply:       true,
		identity.PermissionLicensingRead:      true,
		identity.PermissionLicensingWrite:     true,
		identity.PermissionAuditRead:          true,
		identity.PermissionAuditExport:        true,
	},
	identity.RoleAdministrator: {
		identity.PermissionSystemHealth:      true,
		identity.PermissionAuthSession:       true,
		identity.PermissionUsersRead:         true,
		identity.PermissionUsersWrite:        true,
		identity.PermissionRolesAssign:       true,
		identity.PermissionRolesAssignOwner:  false,
		identity.PermissionSettingsRead:      true,
		identity.PermissionSettingsWrite:     true,
		identity.PermissionSettingsAdminRead: false,
		identity.PermissionStorageRead:       true,
		identity.PermissionStorageWrite:      true,
		identity.PermissionContainersRead:    true,
		identity.PermissionContainersWrite:   true,
		identity.PermissionVMsRead:           true,
		identity.PermissionVMsWrite:          true,
		identity.PermissionNetworkRead:       true,
		identity.PermissionNetworkWrite:      true,
		identity.PermissionUpdatesRead:       true,
		identity.PermissionUpdatesApply:      true,
		identity.PermissionLicensingRead:     true,
		identity.PermissionLicensingWrite:    true,
		identity.PermissionAuditRead:         false,
		identity.PermissionAuditExport:       false,
	},
	identity.RoleOperator: {
		identity.PermissionSystemHealth:      true,
		identity.PermissionAuthSession:       true,
		identity.PermissionUsersRead:         false,
		identity.PermissionUsersWrite:        false,
		identity.PermissionRolesAssign:       false,
		identity.PermissionRolesAssignOwner:  false,
		identity.PermissionSettingsRead:      true,
		identity.PermissionSettingsWrite:     false,
		identity.PermissionSettingsAdminRead: false,
		identity.PermissionStorageRead:       true,
		identity.PermissionStorageWrite:      true,
		identity.PermissionContainersRead:    true,
		identity.PermissionContainersWrite:   true,
		identity.PermissionVMsRead:           true,
		identity.PermissionVMsWrite:          true,
		identity.PermissionNetworkRead:       true,
		identity.PermissionNetworkWrite:      false,
		identity.PermissionUpdatesRead:       true,
		identity.PermissionUpdatesApply:      true,
		identity.PermissionLicensingRead:     true,
		identity.PermissionLicensingWrite:    false,
		identity.PermissionAuditRead:         false,
		identity.PermissionAuditExport:       false,
	},
	identity.RoleUser: {
		identity.PermissionSystemHealth:      true,
		identity.PermissionAuthSession:       true,
		identity.PermissionUsersRead:         false,
		identity.PermissionUsersWrite:        false,
		identity.PermissionRolesAssign:       false,
		identity.PermissionRolesAssignOwner:  false,
		identity.PermissionSettingsRead:      true,
		identity.PermissionSettingsWrite:     false,
		identity.PermissionSettingsAdminRead: false,
		identity.PermissionStorageRead:       true,
		identity.PermissionStorageWrite:      false,
		identity.PermissionContainersRead:    true,
		identity.PermissionContainersWrite:   false,
		identity.PermissionVMsRead:           true,
		identity.PermissionVMsWrite:          false,
		identity.PermissionNetworkRead:       true,
		identity.PermissionNetworkWrite:      false,
		identity.PermissionUpdatesRead:       true,
		identity.PermissionUpdatesApply:      false,
		identity.PermissionLicensingRead:     true,
		identity.PermissionLicensingWrite:    false,
		identity.PermissionAuditRead:         false,
		identity.PermissionAuditExport:       false,
	},
	identity.RoleReadOnly: {
		identity.PermissionSystemHealth:      true,
		identity.PermissionAuthSession:       true,
		identity.PermissionUsersRead:         false,
		identity.PermissionUsersWrite:        false,
		identity.PermissionRolesAssign:       false,
		identity.PermissionRolesAssignOwner:  false,
		identity.PermissionSettingsRead:      true,
		identity.PermissionSettingsWrite:     false,
		identity.PermissionSettingsAdminRead: false,
		identity.PermissionStorageRead:       true,
		identity.PermissionStorageWrite:      false,
		identity.PermissionContainersRead:    true,
		identity.PermissionContainersWrite:   false,
		identity.PermissionVMsRead:           true,
		identity.PermissionVMsWrite:          false,
		identity.PermissionNetworkRead:       true,
		identity.PermissionNetworkWrite:      false,
		identity.PermissionUpdatesRead:       true,
		identity.PermissionUpdatesApply:      false,
		identity.PermissionLicensingRead:     true,
		identity.PermissionLicensingWrite:    false,
		identity.PermissionAuditRead:         false,
		identity.PermissionAuditExport:       false,
	},
}

func TestRolePermissionMatrixMatchesADR(t *testing.T) {
	for role, perms := range expectedMatrix {
		granted := 0
		for _, perm := range identity.AllPermissions {
			want := perms[perm]
			got := identity.Allows(role, perm)
			if got != want {
				t.Fatalf("Allows(%s, %s) = %v, want %v", role, perm, got, want)
			}
			if got {
				granted++
			}
		}
		if len(identity.PermissionsFor(role)) != granted {
			t.Fatalf("PermissionsFor(%s) count mismatch", role)
		}
	}
}

func TestOwnerHasEveryNormativePermission(t *testing.T) {
	for _, perm := range identity.AllPermissions {
		if !identity.Allows(identity.RoleOwner, perm) {
			t.Fatalf("Owner missing permission %s", perm)
		}
	}
}

func TestAdministratorCannotAssignOwner(t *testing.T) {
	if identity.Allows(identity.RoleAdministrator, identity.PermissionRolesAssignOwner) {
		t.Fatal("Administrator must not have roles:assign_owner")
	}
	if !identity.Allows(identity.RoleAdministrator, identity.PermissionRolesAssign) {
		t.Fatal("Administrator must retain roles:assign for non-owner roles")
	}
}

func TestAdministratorDeniedAuditPermissions(t *testing.T) {
	if identity.Allows(identity.RoleAdministrator, identity.PermissionAuditRead) {
		t.Fatal("Administrator must not have audit:read")
	}
	if identity.Allows(identity.RoleAdministrator, identity.PermissionAuditExport) {
		t.Fatal("Administrator must not have audit:export")
	}
}

func TestOwnerAloneGrantsAuditPermissions(t *testing.T) {
	for _, role := range identity.AllRoles {
		if role == identity.RoleOwner {
			continue
		}
		if identity.Allows(role, identity.PermissionAuditRead) {
			t.Fatalf("%s must not have audit:read", role)
		}
		if identity.Allows(role, identity.PermissionAuditExport) {
			t.Fatalf("%s must not have audit:export", role)
		}
	}
	if !identity.Allows(identity.RoleOwner, identity.PermissionAuditRead) {
		t.Fatal("Owner must have audit:read")
	}
	if !identity.Allows(identity.RoleOwner, identity.PermissionAuditExport) {
		t.Fatal("Owner must have audit:export")
	}
}

func TestOperatorDeniedPrivilegedPermissions(t *testing.T) {
	denied := []identity.Permission{
		identity.PermissionUsersRead,
		identity.PermissionUsersWrite,
		identity.PermissionRolesAssign,
		identity.PermissionRolesAssignOwner,
		identity.PermissionSettingsWrite,
		identity.PermissionSettingsAdminRead,
		identity.PermissionSettingsAdminWrite,
		identity.PermissionNetworkWrite,
		identity.PermissionLicensingWrite,
		identity.PermissionAuditRead,
		identity.PermissionAuditExport,
	}
	for _, perm := range denied {
		if identity.Allows(identity.RoleOperator, perm) {
			t.Fatalf("Operator must not have %s", perm)
		}
	}
}

func TestReadOnlyHasNoWriteOrApplyExceptAuthSession(t *testing.T) {
	for _, perm := range identity.AllPermissions {
		if perm == identity.PermissionAuthSession {
			continue
		}
		if !perm.IsMutating() {
			continue
		}
		if identity.Allows(identity.RoleReadOnly, perm) {
			t.Fatalf("Read-only must not have mutating permission %s", perm)
		}
	}
}

func TestUnknownRoleDenied(t *testing.T) {
	role := identity.Role("invalid")
	if role.Valid() {
		t.Fatal("invalid role should not validate")
	}
	if identity.Allows(role, identity.PermissionSystemHealth) {
		t.Fatal("unknown role must be denied")
	}
}

func TestUnknownPermissionDenied(t *testing.T) {
	perm := identity.Permission("unknown:action")
	if perm.Valid() {
		t.Fatal("invalid permission should not validate")
	}
	if identity.Allows(identity.RoleOwner, perm) {
		t.Fatal("unknown permission must be denied even for Owner")
	}
}

func TestOwnerAloneGrantsSettingsAdminRead(t *testing.T) {
	for _, role := range identity.AllRoles {
		if role == identity.RoleOwner {
			continue
		}
		if identity.Allows(role, identity.PermissionSettingsAdminRead) {
			t.Fatalf("%s must not have settings:admin:read", role)
		}
	}
	if !identity.Allows(identity.RoleOwner, identity.PermissionSettingsAdminRead) {
		t.Fatal("Owner must have settings:admin:read")
	}
}

func TestOwnerAloneGrantsSettingsAdminWrite(t *testing.T) {
	for _, role := range identity.AllRoles {
		if role == identity.RoleOwner {
			continue
		}
		if identity.Allows(role, identity.PermissionSettingsAdminWrite) {
			t.Fatalf("%s must not have settings:admin:write", role)
		}
	}
	if !identity.Allows(identity.RoleOwner, identity.PermissionSettingsAdminWrite) {
		t.Fatal("Owner must have settings:admin:write")
	}
}

func TestNoImplicitPrivilegeEscalationAcrossRoles(t *testing.T) {
	roleGrantCounts := map[identity.Role]int{
		identity.RoleOwner:         24,
		identity.RoleAdministrator: 19,
		identity.RoleOperator:      13,
		identity.RoleUser:          9,
		identity.RoleReadOnly:      9,
	}
	for role, wantCount := range roleGrantCounts {
		if got := len(identity.PermissionsFor(role)); got != wantCount {
			t.Fatalf("PermissionsFor(%s) = %d, want %d", role, got, wantCount)
		}
	}
	if len(identity.PermissionsFor(identity.RoleUser)) >= len(identity.PermissionsFor(identity.RoleOperator)) {
		t.Fatal("User must have fewer grants than Operator")
	}
}
