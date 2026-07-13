package identity

// Allows reports whether role grants permission per ADR-0004 first-release matrix.
// Unknown roles and permissions are denied.
func Allows(role Role, perm Permission) bool {
	if !role.Valid() || !perm.Valid() {
		return false
	}
	grants, ok := rolePermissionMatrix[role]
	if !ok {
		return false
	}
	_, ok = grants[perm]
	return ok
}

// PermissionsFor returns a copy of permissions granted to role.
func PermissionsFor(role Role) []Permission {
	if !role.Valid() {
		return nil
	}
	grants := rolePermissionMatrix[role]
	out := make([]Permission, 0, len(grants))
	for _, p := range AllPermissions {
		if _, ok := grants[p]; ok {
			out = append(out, p)
		}
	}
	return out
}

// rolePermissionMatrix is the normative ADR-0004 v1 role/permission matrix.
// Cell R, W, or A grants the permission; — denies it.
var rolePermissionMatrix = map[Role]map[Permission]struct{}{
	RoleOwner: {
		PermissionSystemHealth:       {},
		PermissionAuthSession:        {},
		PermissionUsersRead:          {},
		PermissionUsersWrite:         {},
		PermissionRolesAssign:        {},
		PermissionRolesAssignOwner:   {},
		PermissionSettingsRead:       {},
		PermissionSettingsWrite:      {},
		PermissionSettingsAdminRead:  {},
		PermissionSettingsAdminWrite: {},
		PermissionStorageRead:        {},
		PermissionStorageWrite:       {},
		PermissionContainersRead:     {},
		PermissionContainersWrite:    {},
		PermissionVMsRead:            {},
		PermissionVMsWrite:           {},
		PermissionNetworkRead:        {},
		PermissionNetworkWrite:       {},
		PermissionUpdatesRead:        {},
		PermissionUpdatesApply:       {},
		PermissionLicensingRead:      {},
		PermissionLicensingWrite:     {},
		PermissionAuditRead:          {},
		PermissionAuditExport:        {},
	},
	RoleAdministrator: {
		PermissionSystemHealth:    {},
		PermissionAuthSession:     {},
		PermissionUsersRead:       {},
		PermissionUsersWrite:      {},
		PermissionRolesAssign:     {},
		PermissionSettingsRead:    {},
		PermissionSettingsWrite:   {},
		PermissionStorageRead:     {},
		PermissionStorageWrite:    {},
		PermissionContainersRead:  {},
		PermissionContainersWrite: {},
		PermissionVMsRead:         {},
		PermissionVMsWrite:        {},
		PermissionNetworkRead:     {},
		PermissionNetworkWrite:    {},
		PermissionUpdatesRead:     {},
		PermissionUpdatesApply:    {},
		PermissionLicensingRead:   {},
		PermissionLicensingWrite:  {},
	},
	RoleOperator: {
		PermissionSystemHealth:    {},
		PermissionAuthSession:     {},
		PermissionSettingsRead:    {},
		PermissionStorageRead:     {},
		PermissionStorageWrite:    {},
		PermissionContainersRead:  {},
		PermissionContainersWrite: {},
		PermissionVMsRead:         {},
		PermissionVMsWrite:        {},
		PermissionNetworkRead:     {},
		PermissionUpdatesRead:     {},
		PermissionUpdatesApply:    {},
		PermissionLicensingRead:   {},
	},
	RoleUser: {
		PermissionSystemHealth:   {},
		PermissionAuthSession:    {},
		PermissionSettingsRead:   {},
		PermissionStorageRead:    {},
		PermissionContainersRead: {},
		PermissionVMsRead:        {},
		PermissionNetworkRead:    {},
		PermissionUpdatesRead:    {},
		PermissionLicensingRead:  {},
	},
	RoleReadOnly: {
		PermissionSystemHealth:   {},
		PermissionAuthSession:    {},
		PermissionSettingsRead:   {},
		PermissionStorageRead:    {},
		PermissionContainersRead: {},
		PermissionVMsRead:        {},
		PermissionNetworkRead:    {},
		PermissionUpdatesRead:    {},
		PermissionLicensingRead:  {},
	},
}
