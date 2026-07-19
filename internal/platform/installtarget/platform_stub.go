//go:build !linux

package installtarget

import "errors"

var errPlatformUnsupported = errors.New("installtarget: platform mount unsupported")

func defaultMountController() MountController {
	return unsupportedMountController{}
}

func defaultFSProber() FSProber {
	return unsupportedFSProber{}
}

func defaultPartitionLister() PartitionLister {
	return unsupportedPartitionLister{}
}

type unsupportedMountController struct{}

func (unsupportedMountController) Mount(string, string, string) error {
	return errPlatformUnsupported
}

func (unsupportedMountController) Unmount(string) error {
	return errPlatformUnsupported
}

type unsupportedFSProber struct{}

func (unsupportedFSProber) ProbeFSType(string) (string, error) {
	return "", errPlatformUnsupported
}

type unsupportedPartitionLister struct{}

func (unsupportedPartitionLister) ListPartitions(string) ([]string, error) {
	return nil, errPlatformUnsupported
}
