//go:build linux

package hostmetrics

import (
	"fmt"
	"syscall"
)

func statfsCapacity(stateDir string) (StatFSResult, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(stateDir, &stat); err != nil {
		return StatFSResult{}, err
	}
	if stat.Blocks == 0 {
		return StatFSResult{}, fmt.Errorf("statfs: invalid blocks")
	}

	blockSize := uint64(stat.Bsize)
	totalBytes, err := multiplyUint64(uint64(stat.Blocks), blockSize)
	if err != nil {
		return StatFSResult{}, err
	}
	availableBytes, err := multiplyUint64(uint64(stat.Bavail), blockSize)
	if err != nil {
		return StatFSResult{}, err
	}
	usedBytes, err := DeriveUsedBytes(totalBytes, availableBytes)
	if err != nil {
		return StatFSResult{}, err
	}

	return StatFSResult{
		TotalBytes:     totalBytes,
		AvailableBytes: availableBytes,
		UsedBytes:      usedBytes,
		FSType:         MapFSType(uint32(stat.Type)),
	}, nil
}

func multiplyUint64(a, b uint64) (uint64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	if a > ^uint64(0)/b {
		return 0, fmt.Errorf("statfs: overflow")
	}
	return a * b, nil
}
