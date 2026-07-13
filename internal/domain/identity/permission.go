package identity

import "fmt"

// Permission is a normative v1 resource:action grant (ADR-0004).
type Permission string

const (
	PermissionSystemHealth       Permission = "system:health"
	PermissionAuthSession        Permission = "auth:session"
	PermissionUsersRead          Permission = "users:read"
	PermissionUsersWrite         Permission = "users:write"
	PermissionRolesAssign        Permission = "roles:assign"
	PermissionRolesAssignOwner   Permission = "roles:assign_owner"
	PermissionSettingsRead       Permission = "settings:read"
	PermissionSettingsWrite      Permission = "settings:write"
	PermissionSettingsAdminRead  Permission = "settings:admin:read"
	PermissionSettingsAdminWrite Permission = "settings:admin:write"
	PermissionStorageRead        Permission = "storage:read"
	PermissionStorageWrite       Permission = "storage:write"
	PermissionContainersRead     Permission = "containers:read"
	PermissionContainersWrite    Permission = "containers:write"
	PermissionVMsRead            Permission = "vms:read"
	PermissionVMsWrite           Permission = "vms:write"
	PermissionNetworkRead        Permission = "network:read"
	PermissionNetworkWrite       Permission = "network:write"
	PermissionUpdatesRead        Permission = "updates:read"
	PermissionUpdatesApply       Permission = "updates:apply"
	PermissionLicensingRead      Permission = "licensing:read"
	PermissionLicensingWrite     Permission = "licensing:write"
	PermissionAuditRead          Permission = "audit:read"
	PermissionAuditExport        Permission = "audit:export"
)

// AllPermissions lists every normative v1 permission in stable order.
var AllPermissions = []Permission{
	PermissionSystemHealth,
	PermissionAuthSession,
	PermissionUsersRead,
	PermissionUsersWrite,
	PermissionRolesAssign,
	PermissionRolesAssignOwner,
	PermissionSettingsRead,
	PermissionSettingsWrite,
	PermissionSettingsAdminRead,
	PermissionSettingsAdminWrite,
	PermissionStorageRead,
	PermissionStorageWrite,
	PermissionContainersRead,
	PermissionContainersWrite,
	PermissionVMsRead,
	PermissionVMsWrite,
	PermissionNetworkRead,
	PermissionNetworkWrite,
	PermissionUpdatesRead,
	PermissionUpdatesApply,
	PermissionLicensingRead,
	PermissionLicensingWrite,
	PermissionAuditRead,
	PermissionAuditExport,
}

// Valid reports whether permission is a known normative v1 permission.
func (p Permission) Valid() bool {
	_, ok := permissionIndex[p]
	return ok
}

// String returns the permission identifier.
func (p Permission) String() string {
	return string(p)
}

// ParsePermission parses a normative permission identifier.
func ParsePermission(s string) (Permission, error) {
	p := Permission(s)
	if !p.Valid() {
		return "", fmt.Errorf("%w: %q", ErrInvalidPermission, s)
	}
	return p, nil
}

// IsMutating reports whether the permission implies write or apply capability.
func (p Permission) IsMutating() bool {
	switch p {
	case PermissionUsersWrite,
		PermissionRolesAssign,
		PermissionRolesAssignOwner,
		PermissionSettingsWrite,
		PermissionSettingsAdminWrite,
		PermissionStorageWrite,
		PermissionContainersWrite,
		PermissionVMsWrite,
		PermissionNetworkWrite,
		PermissionUpdatesApply,
		PermissionLicensingWrite,
		PermissionAuditExport,
		PermissionAuthSession:
		return true
	default:
		return false
	}
}

var permissionIndex = func() map[Permission]struct{} {
	m := make(map[Permission]struct{}, len(AllPermissions))
	for _, p := range AllPermissions {
		m[p] = struct{}{}
	}
	return m
}()
