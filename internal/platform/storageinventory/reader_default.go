package storageinventory

// defaultBlockDeviceReader is resolved at build time via platform-specific files.
func defaultBlockDeviceReader() BlockDeviceReader {
	return platformBlockDeviceReader()
}
