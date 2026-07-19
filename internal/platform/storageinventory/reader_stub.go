//go:build !linux

package storageinventory

type stubBlockDeviceReader struct{}

func platformBlockDeviceReader() BlockDeviceReader {
	return stubBlockDeviceReader{}
}

func (stubBlockDeviceReader) ListBlockDevices(string) ([]RawDevice, error) {
	return nil, errStorageDiscoveryUnavailable
}
