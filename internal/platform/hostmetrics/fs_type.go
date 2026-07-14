package hostmetrics

const (
	fsTypeExt4  = "ext4"
	fsTypeXFS   = "xfs"
	fsTypeBtrfs = "btrfs"
	fsTypeTmpfs = "tmpfs"
	fsTypeOther = "other"
)

// MapFSType maps a statfs type magic number to a small fixed vocabulary.
func MapFSType(magic uint32) string {
	switch magic {
	case 0xEF53:
		return fsTypeExt4
	case 0x58465342:
		return fsTypeXFS
	case 0x9123683E:
		return fsTypeBtrfs
	case 0x01021994:
		return fsTypeTmpfs
	default:
		return fsTypeOther
	}
}
